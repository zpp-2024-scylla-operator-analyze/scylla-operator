// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/scylladb/scylla-operator/pkg/mermaidclient/internal/models"
)

// PutClusterClusterIDTaskTaskTypeTaskIDStartReader is a Reader for the PutClusterClusterIDTaskTaskTypeTaskIDStart structure.
type PutClusterClusterIDTaskTaskTypeTaskIDStartReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPutClusterClusterIDTaskTaskTypeTaskIDStartOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewPutClusterClusterIDTaskTaskTypeTaskIDStartDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewPutClusterClusterIDTaskTaskTypeTaskIDStartOK creates a PutClusterClusterIDTaskTaskTypeTaskIDStartOK with default headers values
func NewPutClusterClusterIDTaskTaskTypeTaskIDStartOK() *PutClusterClusterIDTaskTaskTypeTaskIDStartOK {
	return &PutClusterClusterIDTaskTaskTypeTaskIDStartOK{}
}

/*PutClusterClusterIDTaskTaskTypeTaskIDStartOK handles this case with default header values.

Task started
*/
type PutClusterClusterIDTaskTaskTypeTaskIDStartOK struct {
}

func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartOK) Error() string {
	return fmt.Sprintf("[PUT /cluster/{cluster_id}/task/{task_type}/{task_id}/start][%d] putClusterClusterIdTaskTaskTypeTaskIdStartOK ", 200)
}

func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewPutClusterClusterIDTaskTaskTypeTaskIDStartDefault creates a PutClusterClusterIDTaskTaskTypeTaskIDStartDefault with default headers values
func NewPutClusterClusterIDTaskTaskTypeTaskIDStartDefault(code int) *PutClusterClusterIDTaskTaskTypeTaskIDStartDefault {
	return &PutClusterClusterIDTaskTaskTypeTaskIDStartDefault{
		_statusCode: code,
	}
}

/*PutClusterClusterIDTaskTaskTypeTaskIDStartDefault handles this case with default header values.

Unexpected error
*/
type PutClusterClusterIDTaskTaskTypeTaskIDStartDefault struct {
	_statusCode int

	Payload *models.ErrorResponse
}

// Code gets the status code for the put cluster cluster ID task task type task ID start default response
func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartDefault) Code() int {
	return o._statusCode
}

func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartDefault) Error() string {
	return fmt.Sprintf("[PUT /cluster/{cluster_id}/task/{task_type}/{task_id}/start][%d] PutClusterClusterIDTaskTaskTypeTaskIDStart default  %+v", o._statusCode, o.Payload)
}

func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartDefault) GetPayload() *models.ErrorResponse {
	return o.Payload
}

func (o *PutClusterClusterIDTaskTaskTypeTaskIDStartDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorResponse)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
