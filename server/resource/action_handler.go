package resource

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/artpar/api2go"
	"github.com/daptin/daptin/server/auth"
	"github.com/dop251/goja"
	log "github.com/sirupsen/logrus"
	"gopkg.in/gin-gonic/gin.v1"
	//"io"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"

	"github.com/artpar/conform"
	english "github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"gopkg.in/go-playground/validator.v9"
	en2 "gopkg.in/go-playground/validator.v9/translations/en"
)

var guestActions = map[string]Action{}

func CreateGuestActionListHandler(initConfig *CmsConfig, cruds map[string]*DbResource) func(*gin.Context) {

	actionMap := make(map[string]Action)

	for _, ac := range initConfig.Actions {
		actionMap[ac.OnType+":"+ac.Name] = ac
	}

	guestActions["user:signup"] = actionMap["user:signup"]
	guestActions["user:signin"] = actionMap["user:signin"]

	return func(c *gin.Context) {

		c.JSON(200, guestActions)
	}
}

func CreateGetActionHandler(initConfig *CmsConfig, configStore *ConfigStore, cruds map[string]*DbResource) func(*gin.Context) {
	return func(ginContext *gin.Context) {

	}
}

type ActionPerformerInterface interface {
	DoAction(request ActionRequest, inFields map[string]interface{}) ([]ActionResponse, []error)
	Name() string
}

type DaptinError struct {
	Message string
	Code    string
}

func NewDaptinError(str string, code string) DaptinError {
	return DaptinError{
		Message: str,
		Code:    code,
	}
}

func CreatePostActionHandler(initConfig *CmsConfig, configStore *ConfigStore, cruds map[string]*DbResource, actionPerformers []ActionPerformerInterface) func(*gin.Context) {

	actionMap := make(map[string]Action)

	eng := english.New()
	uni := ut.New(eng, eng)
	trans, _ := uni.GetTranslator("en")

	err := en2.RegisterDefaultTranslations(initConfig.Validator, trans)
	if err != nil {
		log.Errorf("Failed to register translations: %v", err)
	}

	for _, ac := range initConfig.Actions {
		actionMap[ac.OnType+":"+ac.Name] = ac
	}

	actionHandlerMap := make(map[string]ActionPerformerInterface)

	for _, actionPerformer := range actionPerformers {
		if actionPerformer == nil {
			continue
		}
		actionHandlerMap[actionPerformer.Name()] = actionPerformer
	}

	return func(ginContext *gin.Context) {

		actionName := ginContext.Param("actionName")
		//log.Infof("Action name: %v", actionName)

		bytes, err := ioutil.ReadAll(ginContext.Request.Body)
		if err != nil {
			ginContext.Error(err)
			return
		}

		actionRequest := ActionRequest{}
		json.Unmarshal(bytes, &actionRequest)

		actionRequest.Type = ginContext.Param("typename")
		actionRequest.Action = actionName

		if actionRequest.Attributes == nil {
			actionRequest.Attributes = make(map[string]interface{})
		}

		params := ginContext.Params
		for _, param := range params {
			actionRequest.Attributes[param.Key] = param.Value
		}

		//log.Infof("Request body: %v", actionRequest)

		req := api2go.Request{
			PlainRequest: &http.Request{
				Method: "GET",
			},
		}

		req.PlainRequest = req.PlainRequest.WithContext(ginContext.Request.Context())

		user := ginContext.Request.Context().Value("user")
		var sessionUser auth.SessionUser

		if user != nil {
			sessionUser = user.(auth.SessionUser)
		}

		var subjectInstance *api2go.Api2GoModel
		var subjectInstanceMap map[string]interface{}

		subjectInstanceReferenceId, ok := actionRequest.Attributes[actionRequest.Type+"_id"]
		if ok {
			referencedObject, err := cruds[actionRequest.Type].FindOne(subjectInstanceReferenceId.(string), req)
			if err != nil {
				ginContext.AbortWithError(400, err)
				return
			}
			subjectInstance = referencedObject.Result().(*api2go.Api2GoModel)

			subjectInstanceMap = subjectInstance.Data

			if subjectInstanceMap == nil {
				ginContext.AbortWithError(403, errors.New("Forbidden"))
				return
			}

			subjectInstanceMap["__type"] = subjectInstance.GetName()
			permission := cruds[actionRequest.Type].GetRowPermission(subjectInstanceMap)

			if !permission.CanExecute(sessionUser.UserReferenceId, sessionUser.Groups) {
				ginContext.AbortWithError(403, errors.New("Forbidden"))
				return
			}
		}

		if !cruds["world"].IsUserActionAllowed(sessionUser.UserReferenceId, sessionUser.Groups, actionRequest.Type, actionRequest.Action) {
			ginContext.AbortWithError(403, errors.New("Forbidden"))
			return
		}

		log.Infof("Handle event for action [%v]", actionName)

		action, err := cruds["action"].GetActionByName(actionRequest.Type, actionRequest.Action)

		for _, field := range action.InFields {
			_, ok := actionRequest.Attributes[field.ColumnName]
			if !ok {
				actionRequest.Attributes[field.ColumnName] = ginContext.Query(field.ColumnName)
			}
		}

		for _, validation := range action.Validations {
			errs := initConfig.Validator.VarWithValue(actionRequest.Attributes[validation.ColumnName], actionRequest.Attributes, validation.Tags)
			if errs != nil {
				validationErrors := errs.(validator.ValidationErrors)
				firstError := validationErrors[0]
				ginContext.JSON(400, NewDaptinError(validation.ColumnName+": "+firstError.Translate(trans), "validation-failed"))
				//ginContext.AbortWithError(400, errors.New(validationErrors[0].Translate(en1)))
				return
			}
		}

		for _, conformations := range action.Conformations {

			val, ok := actionRequest.Attributes[conformations.ColumnName]
			if !ok {
				continue
			}
			valStr, ok := val.(string)
			if !ok {
				continue
			}
			newVal := conform.TransformString(valStr, conformations.Tags)
			actionRequest.Attributes[conformations.ColumnName] = newVal
		}

		inFieldMap, err := GetValidatedInFields(actionRequest, action)

		if err != nil {
			ginContext.AbortWithError(400, err)
			return
		}

		if sessionUser.UserReferenceId != "" {
			user, err := cruds["user"].GetReferenceIdToObject("user", sessionUser.UserReferenceId)
			if err != nil {
				log.Errorf("Failed to load user: %v", err)
				return
			}
			inFieldMap["user"] = user
		}

		if subjectInstanceMap != nil {
			inFieldMap[actionRequest.Type+"_id"] = subjectInstanceMap["reference_id"]
			inFieldMap["subject"] = subjectInstanceMap
		}

		responses := make([]ActionResponse, 0)

	OutFields:
		for _, outcome := range action.OutFields {
			var res api2go.Responder
			var responses1 []ActionResponse
			var errors1 []error
			var actionResponse ActionResponse

			model, request, err := BuildOutcome(inFieldMap, outcome)
			if err != nil {
				log.Errorf("Failed to build outcome: %v", err)
				responses = append(responses, NewActionResponse("error", "Failed to build outcome "+outcome.Type))
				continue
			}

			request.PlainRequest = request.PlainRequest.WithContext(ginContext.Request.Context())

			dbResource, ok := cruds[outcome.Type]
			if !ok {
				//log.Errorf("No DbResource for type [%v]", outcome.Type)
			}

			log.Infof("Next outcome method: %v", outcome.Method)
			switch outcome.Method {
			case "POST":
				res, err = dbResource.Create(model, request)
				if err != nil {

					actionResponse = NewActionResponse("client.notify", NewClientNotification("error", "Failed to create "+model.GetName()+". "+err.Error(), "Failed"))
					responses = append(responses, actionResponse)
					break OutFields
				} else {
					actionResponse = NewActionResponse("client.notify", NewClientNotification("success", "Created "+model.GetName(), "Success"))
				}
				responses = append(responses, actionResponse)
			case "UPDATE":
				res, err = dbResource.Update(model, request)
				if err != nil {
					actionResponse = NewActionResponse("client.notify", NewClientNotification("error", "Failed to update "+model.GetName()+". "+err.Error(), "Failed"))
					responses = append(responses, actionResponse)
					break OutFields
				} else {
					actionResponse = NewActionResponse("client.notify", NewClientNotification("success", "Created "+model.GetName(), "Success"))
				}
				responses = append(responses, actionResponse)
			case "DELETE":
				res, err = dbResource.Delete(model.Data["reference_id"].(string), request)
				if err != nil {
					actionResponse = NewActionResponse("client.notify", NewClientNotification("error", "Failed to delete "+model.GetName(), "Failed"))
					responses = append(responses, actionResponse)
					break OutFields
				} else {
					actionResponse = NewActionResponse("client.notify", NewClientNotification("success", "Created "+model.GetName(), "Success"))
				}
				responses = append(responses, actionResponse)
			case "EXECUTE":
				//res, err = cruds[outcome.Type].Create(model, request)

				performer, ok := actionHandlerMap[model.GetName()]
				if !ok {
					log.Errorf("Invalid outcome method: [%v]%v", outcome.Method, model.GetName())
					//return ginContext.AbortWithError(500, errors.New("Invalid outcome"))
				} else {
					responses1, errors1 = performer.DoAction(actionRequest, model.Data)
					responses = append(responses, responses1...)
					if errors1 != nil && len(errors1) > 0 {
						err = errors1[0]
					}
				}

			case "ACTIONRESPONSE":
				//res, err = cruds[outcome.Type].Create(model, request)
				log.Infof("Create action response: ", model.GetName())
				var actionResponse ActionResponse
				switch model.GetName() {
				case "client.notify":
					actionResponse = NewActionResponse("client.notify", model.Data)
				case "client.redirect":
					actionResponse = NewActionResponse("client.redirect", model.Data)
				case "client.store.set":
					actionResponse = NewActionResponse("client.store.set", model.Data)
				case "error":
					actionResponse = NewActionResponse("error", model.Data)
				default:
					log.Errorf("Unknown action response type: %v", model.GetName())
				}
				responses = append(responses, actionResponse)

			default:
				log.Errorf("Unknown outcome method: %v", outcome.Method)

			}
			if res != nil && res.Result() != nil {
				inFieldMap[outcome.Reference] = res.Result().(*api2go.Api2GoModel).Data

				if outcome.Reference == "" {
					inFieldMap["subject"] = inFieldMap[outcome.Reference]
				}
			}

			if len(responses1) > 0 {
				lst := make([]interface{}, 0)
				for i, res := range responses1 {
					inFieldMap[fmt.Sprintf("%v[%v]", outcome.Reference, i)] = res.Attributes
					lst = append(lst, res.Attributes)
				}
				inFieldMap[outcome.Reference] = lst
			}

			if err != nil {
				ginContext.AbortWithError(500, err)
				return
			}
		}

		//log.Infof("Final responses: %v", responses)

		ginContext.JSON(200, responses)

	}
}
func NewClientNotification(notificationType string, message string, title string) map[string]interface{} {

	m := make(map[string]interface{})

	m["type"] = notificationType
	m["message"] = message
	m["title"] = title
	return m

}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

type ActionResponse struct {
	ResponseType string
	Attributes   interface{}
}

func NewActionResponse(responseType string, attrs interface{}) ActionResponse {

	ar := ActionResponse{
		ResponseType: responseType,
		Attributes:   attrs,
	}

	return ar

}

func BuildOutcome(inFieldMap map[string]interface{}, outcome Outcome) (*api2go.Api2GoModel, api2go.Request, error) {

	attrInterface, err := buildActionContext(outcome.Attributes, inFieldMap)
	if err != nil {
		return nil, api2go.Request{}, err
	}
	attrs := attrInterface.(map[string]interface{})

	switch outcome.Type {
	case "system_json_schema_update":
		responseModel := api2go.NewApi2GoModel("__restart", nil, 0, nil)
		returnRequest := api2go.Request{
			PlainRequest: &http.Request{
				Method: "EXECUTE",
			},
		}

		files1, ok := attrs["json_schema"]
		log.Infof("Files [%v]: %v", attrs, files1)
		files := files1.([]interface{})
		if !ok || len(files) < 1 {
			return nil, returnRequest, errors.New("No files uploaded")
		}
		for _, file := range files {
			f := file.(map[string]interface{})
			fileName := f["name"].(string)
			log.Infof("File name: %v", fileName)
			fileNameParts := strings.Split(fileName, ".")
			fileFormat := fileNameParts[len(fileNameParts)-1]
			fileContentsBase64 := f["file"].(string)
			fileBytes, err := base64.StdEncoding.DecodeString(strings.Split(fileContentsBase64, ",")[1])
			if err != nil {
				return nil, returnRequest, err
			}

			jsonFileName := fmt.Sprintf("schema_%v_daptin.%v", fileName, fileFormat)
			err = ioutil.WriteFile(jsonFileName, fileBytes, 0644)
			if err != nil {
				log.Errorf("Failed to write json file: %v", jsonFileName)
				return nil, returnRequest, err
			}

		}

		log.Infof("Written all json files. Attempting restart")

		return responseModel, returnRequest, nil

	case "__download_cms_config":
		fallthrough
	case "__become_admin":

		returnRequest := api2go.Request{
			PlainRequest: &http.Request{
				Method: "EXECUTE",
			},
		}
		model := api2go.NewApi2GoModelWithData(outcome.Type, nil, auth.DEFAULT_PERMISSION.IntValue(), nil, attrs)

		return model, returnRequest, nil

	case "action.response":
		fallthrough
	case "client.redirect":
		fallthrough
	case "client.store.set":
		fallthrough
	case "client.notify":
		//respopnseModel := NewActionResponse(attrs["responseType"].(string), attrs)
		returnRequest := api2go.Request{
			PlainRequest: &http.Request{
				Method: "ACTIONRESPONSE",
			},
		}
		model := api2go.NewApi2GoModelWithData(outcome.Type, nil, auth.DEFAULT_PERMISSION.IntValue(), nil, attrs)

		return model, returnRequest, nil

	case "POST":

		model := api2go.NewApi2GoModelWithData(outcome.Type, nil, auth.DEFAULT_PERMISSION.IntValue(), nil, attrs)

		req := api2go.Request{
			PlainRequest: &http.Request{
				Method: "POST",
			},
		}
		return model, req, nil
	case "UPDATE":

		model := api2go.NewApi2GoModelWithData(outcome.Type, nil, auth.DEFAULT_PERMISSION.IntValue(), nil, attrs)

		req := api2go.Request{
			PlainRequest: &http.Request{
				Method: "POST",
			},
		}
		return model, req, nil
	default:

		model := api2go.NewApi2GoModelWithData(outcome.Type, nil, auth.DEFAULT_PERMISSION.IntValue(), nil, attrs)

		req := api2go.Request{
			PlainRequest: &http.Request{
				Method: "POST",
			},
		}
		return model, req, nil

	}

	//return nil, api2go.Request{}, errors.New(fmt.Sprintf("Unidentified outcome: %v", outcome.Type))

}

func runUnsafeJavascript(unsafe string, contextMap map[string]interface{}) (interface{}, error) {

	vm := goja.New()

	//vm.ToValue(contextMap)
	for key, val := range contextMap {
		vm.Set(key, val)
	}
	v, err := vm.RunString(unsafe) // Here be dragons (risky code)

	if err != nil {
		return nil, err
	}

	return v.Export(), nil
}

func buildActionContext(outcomeAttributes interface{}, inFieldMap map[string]interface{}) (interface{}, error) {

	var data interface{}

	kindOfOutcome := reflect.TypeOf(outcomeAttributes).Kind()

	if kindOfOutcome == reflect.Map {

		dataMap := make(map[string]interface{})

		outcomeMap := outcomeAttributes.(map[string]interface{})
		for key, field := range outcomeMap {

			typeOfField := reflect.TypeOf(field).Kind()
			//log.Infof("Outcome attribute [%v] == %v [%v]", key, field, typeOfField)

			if typeOfField == reflect.String {

				fieldString := field.(string)

				val, err := evaluateString(fieldString, inFieldMap)
				if err != nil {
					return nil, err
				}
				dataMap[key] = val

			} else if typeOfField == reflect.Map || typeOfField == reflect.Slice || typeOfField == reflect.Array {

				val, err := buildActionContext(field, inFieldMap)
				if err != nil {
					return nil, err
				}
				dataMap[key] = val

			}

		}

		data = dataMap

	} else if kindOfOutcome == reflect.Array || kindOfOutcome == reflect.Slice {

		outcomeArray, ok := outcomeAttributes.([]interface{})

		if !ok {
			outcomeArray = make([]interface{}, 0)
			outcomeArrayString := outcomeAttributes.([]string)
			for _, o := range outcomeArrayString {
				outcomeArray = append(outcomeArray, o)
			}
		}

		outcomes := make([]interface{}, 0)

		for _, outcome := range outcomeArray {

			outcomeKind := reflect.TypeOf(outcome).Kind()

			if outcomeKind == reflect.String {

				outcomeString := outcome.(string)

				evtStr, err := evaluateString(outcomeString, inFieldMap)
				if err != nil {
					return data, err
				}
				outcomes = append(outcomes, evtStr)

			} else if outcomeKind == reflect.Map || outcomeKind == reflect.Array || outcomeKind == reflect.Slice {
				outc, err := buildActionContext(outcome, inFieldMap)
				if err != nil {
					return data, err
				}
				outcomes = append(outcomes, outc)
			}

		}
		data = outcomes

	}

	return data, nil
}

func evaluateString(fieldString string, inFieldMap map[string]interface{}) (interface{}, error) {

	var val interface{}

	if fieldString == "" {
		return "", nil
	}

	if fieldString[0] == '!' {

		res, err := runUnsafeJavascript(fieldString[1:], inFieldMap)
		if err != nil {
			return nil, err
		}
		val = res

	} else if fieldString[0] == '~' {

		fieldParts := strings.Split(fieldString[1:], ".")

		if fieldParts[0] == "" {
			fieldParts[0] = "subject"
		}
		var finalValue interface{}

		// it looks confusing but it does whats its supposed to do
		// todo: add helpful comment

		finalValue = inFieldMap
		for i := 0; i < len(fieldParts)-1; i++ {
			fieldPart := fieldParts[i]
			finalValue = finalValue.(map[string]interface{})[fieldPart]
		}
		if finalValue == nil {
			return nil, nil
		}

		castMap := finalValue.(map[string]interface{})
		finalValue = castMap[fieldParts[len(fieldParts)-1]]
		val = finalValue

	} else {

		rex := regexp.MustCompile(`\$([a-zA-Z0-9_\[\]]+)?(\.[a-zA-Z0-9_]+)+`)
		matches := rex.FindAllStringSubmatch(fieldString, -1)

		for _, match := range matches {

			fieldParts := strings.Split(match[0][1:], ".")

			if fieldParts[0] == "" {
				fieldParts[0] = "subject"
			}

			var finalValue interface{}

			// it looks confusing but it does whats its supposed to do
			// todo: add helpful comment

			finalValue = inFieldMap
			for i := 0; i < len(fieldParts)-1; i++ {
				fieldPart := fieldParts[i]
				finalValue = finalValue.(map[string]interface{})[fieldPart]
			}
			if finalValue == nil {
				return nil, nil
			}

			castMap := finalValue.(map[string]interface{})
			finalValue = castMap[fieldParts[len(fieldParts)-1]]
			fieldString = strings.Replace(fieldString, fmt.Sprintf("%v", match[0]), fmt.Sprintf("%v", finalValue), -1)
		}
		val = fieldString

	}

	return val, nil

}

func GetValidatedInFields(actionRequest ActionRequest, action Action) (map[string]interface{}, error) {

	dataMap := actionRequest.Attributes
	finalDataMap := make(map[string]interface{})
	for _, inField := range action.InFields {
		val, ok := dataMap[inField.ColumnName]
		if ok {
			finalDataMap[inField.ColumnName] = val

		} else if inField.DefaultValue != "" {
		} else if inField.IsNullable {

		} else {
			return nil, errors.New(fmt.Sprintf("Field %s cannot be blank", inField.Name))
		}
	}

	return finalDataMap, nil
}
