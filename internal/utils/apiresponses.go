package utils

/*
 * File: internal/utils/apiresponses.go
 *
 * Purpose: API layer handlers for API response structuring and construction
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

// Type `tournabyteWebAPIResponse` contains the generalized structure of an API response from the Tournabyte webapi service
//
// Struct members:
//   - Ok: indicates the status of the response
//   - Data: the payload of the successful run (should only exist when ok=true and error=nil)
//   - Error: the payload of the failed run (should only exist when ok=false and data=nil)
//
// Encoding:
//
//	The structure should be encoded to JSON in one of these two variants
//
//	OK variant
//	{
//		"ok": true,
//		"data": ...
//	}
//
//	Error variant
//	{
//		"ok": false,
//		"error" ...
//	}
type tournabyteWebAPIResponse struct {
	Ok    bool  `json:"ok"`
	Data  any   `json:"data,omitempty"`
	Error error `json:"error,omitempty"`
}

// Function `WriteDataResponse` creates a tournabyteWebAPIResponse indicating a successful processing of a request with the response payload
//
// Parameters:
//   - data: the data to include in the response payload
//
// Returns:
//   - `tournabyteWebAPIResponse`: a properly structured response indicating a request was successfully processed
func WriteDataResponse(data any) tournabyteWebAPIResponse {
	return tournabyteWebAPIResponse{
		Ok:    true,
		Data:  data,
		Error: nil,
	}
}

// Function `WriteErrorResponse` creates a tournabyteWebAPIResponse indicating a request was not processed properly with the error payload
//
// Parameters:
//   - err: the error emitted from the service layer
//   - additionalInfo: any relevant information for the error that occurred
//
// Returns:
//   - `tournabyteWebAPIResponse`: a properly structured response indicating an error occurred during request processing
func WriteErrorResponse(err error) tournabyteWebAPIResponse {
	return tournabyteWebAPIResponse{
		Ok:    false,
		Data:  nil,
		Error: err,
	}
}
