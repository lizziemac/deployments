// Copyright 2019 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/ant0ine/go-json-rest/rest/test"

	"github.com/mendersoftware/deployments/app"
	app_mocks "github.com/mendersoftware/deployments/app/mocks"
	store_mocks "github.com/mendersoftware/deployments/store/mocks"
	"github.com/mendersoftware/deployments/utils/restutil/view"
	h "github.com/mendersoftware/deployments/utils/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPostArtifactsGenerate(t *testing.T) {
	type request struct {
		Name                  string `json:"name"`
		Description           string `json:"description"`
		Size                  int64  `json:"size"`
		DeviceTypesCompatible string `json:"device_types_compatible"`
		Type                  string `json:"type"`
		Args                  string `json:"args"`
	}

	imageBody := []byte("123456790")

	testCases := []struct {
		requestBodyObject        []h.Part
		requestContentType       string
		responseCode             int
		responseBody             string
		appGenerateImage         bool
		appGenerateImageResponse string
		appGenerateImageError    error
	}{
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "",
			responseCode:       http.StatusBadRequest,
			responseBody:       "mime: no media type",
		},
		{
			requestBodyObject:  []h.Part{},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "Request does not contain artifact",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "No size provided before the file part of the message or the size value is wrong.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(-1),
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "No size provided before the file part of the message or the size value is wrong.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(0),
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "No size provided before the file part of the message or the size value is wrong.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:   "file",
					ContentType: "",
					ImageData:   imageBody,
				},
			},
			requestContentType: "multipart/form-data",
			responseCode:       http.StatusBadRequest,
			responseBody:       "The last part of the multipart/form-data message should be a file.",
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusCreated,
			responseBody:             "",
			appGenerateImage:         true,
			appGenerateImageResponse: "artifactID",
			appGenerateImageError:    nil,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusUnprocessableEntity,
			responseBody:             "Artifact not unique",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    app.ErrModelArtifactNotUnique,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusBadRequest,
			responseBody:             "Artifact file too large",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    app.ErrModelArtifactFileTooLarge,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusBadRequest,
			responseBody:             "Cannot parse artifact file",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    app.ErrModelParsingArtifactFailed,
		},
		{
			requestBodyObject: []h.Part{
				{
					FieldName:  "name",
					FieldValue: "name",
				},
				{
					FieldName:  "description",
					FieldValue: "description",
				},
				{
					FieldName:  "size",
					FieldValue: strconv.Itoa(len(imageBody)),
				},
				{
					FieldName:  "device_types_compatible",
					FieldValue: "Beagle Bone",
				},
				{
					FieldName:  "type",
					FieldValue: "single_file",
				},
				{
					FieldName:  "args",
					FieldValue: "args",
				},
				{
					FieldName:   "file",
					ContentType: "application/octet-stream",
					ImageData:   imageBody,
				},
			},
			requestContentType:       "multipart/form-data",
			responseCode:             http.StatusInternalServerError,
			responseBody:             "internal error",
			appGenerateImage:         true,
			appGenerateImageResponse: "",
			appGenerateImageError:    errors.New("generic error"),
		},
	}

	store := &store_mocks.DataStore{}
	restView := new(view.RESTView)

	for i := range testCases {
		tc := testCases[i]

		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			app := &app_mocks.App{}

			if tc.appGenerateImage {
				app.On("GenerateImage",
					h.ContextMatcher(),
					mock.AnythingOfType("*model.MultipartGenerateImageMsg"),
				).Return(tc.appGenerateImageResponse, tc.appGenerateImageError)
			}

			d := NewDeploymentsApiHandlers(store, restView, app)
			api := setUpRestTest("/api/0.0.1/artifacts/generate", rest.Post, d.GenerateImage)
			req := h.MakeMultipartRequest("POST", "http://localhost/api/0.0.1/artifacts/generate",
				tc.requestContentType, tc.requestBodyObject)

			recorded := test.RunRequest(t, api.MakeHandler(), req)
			recorded.CodeIs(tc.responseCode)
			if tc.responseBody == "" {
				recorded.BodyIs(tc.responseBody)
			} else {
				body, _ := recorded.DecodedBody()
				assert.Contains(t, string(body), tc.responseBody)
			}

			if tc.appGenerateImage {
				app.AssertExpectations(t)
			}
		})
	}

}
