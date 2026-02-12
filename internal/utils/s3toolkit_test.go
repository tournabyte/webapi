package utils

/*
 * File: internal/utils/s3toolkit_test.go
 *
 * Purpose: unit tests for the s3 toolkit
 *
 * License:
 *  See LICENSE.md for full license
 *  Copyright 2026 Part of the Tournabyte project
 *
 */

import (
	"testing"

	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
)

func TestMinioStaticCredentials(t *testing.T) {
	t.Parallel()

	accessKey := "test-access-key"
	secretKey := "test-secret-key"

	opt := MinioStaticCredentials(accessKey, secretKey)

	// Create a mock options struct
	mockOptions := &minio.Options{}

	// Apply the option
	err := opt(mockOptions)

	assert.NoError(t, err)

	// Verify that credentials were set correctly
	assert.NotNil(t, mockOptions.Creds)

	// Test with empty credentials
	opt2 := MinioStaticCredentials("", "")
	err2 := opt2(mockOptions)

	assert.NoError(t, err2)
	assert.NotNil(t, mockOptions.Creds)
}

func TestMinioUseSecureConnection(t *testing.T) {
	t.Parallel()

	// Test with secure = true
	opt := MinioUseSecureConnection(true)
	mockOptions := &minio.Options{}

	err := opt(mockOptions)
	assert.NoError(t, err)
	assert.True(t, mockOptions.Secure)

	// Test with secure = false
	opt2 := MinioUseSecureConnection(false)
	err2 := opt2(mockOptions)
	assert.NoError(t, err2)
	assert.False(t, mockOptions.Secure)
}

func TestMinioMaxRetries(t *testing.T) {
	t.Parallel()

	retryLimit := 5
	opt := MinioMaxRetries(retryLimit)
	mockOptions := &minio.Options{}

	err := opt(mockOptions)
	assert.NoError(t, err)
	assert.Equal(t, retryLimit, mockOptions.MaxRetries)

	// Test with zero retries
	opt2 := MinioMaxRetries(0)
	err2 := opt2(mockOptions)
	assert.NoError(t, err2)
	assert.Equal(t, 0, mockOptions.MaxRetries)
}

func TestPutObjectMetadata(t *testing.T) {
	t.Parallel()

	// Test with valid metadata pairs
	opt := PutObjectMetadata("key1", "value1", "key2", "value2")
	mockOptions := &minio.PutObjectOptions{}

	err := opt(mockOptions)
	assert.NoError(t, err)

	// Check that metadata was set correctly
	assert.NotNil(t, mockOptions.UserMetadata)
	assert.Equal(t, "value1", mockOptions.UserMetadata["key1"])
	assert.Equal(t, "value2", mockOptions.UserMetadata["key2"])

	// Test with invalid metadata (odd number of arguments)
	opt2 := PutObjectMetadata("key1", "value1", "key2")
	err2 := opt2(mockOptions)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "received odd length input")
}

func TestPutObjectTags(t *testing.T) {
	t.Parallel()

	// Test with valid tag pairs
	opt := PutObjectTags("tag1", "value1", "tag2", "value2")
	mockOptions := &minio.PutObjectOptions{}

	err := opt(mockOptions)
	assert.NoError(t, err)

	// Check that tags were set correctly
	assert.NotNil(t, mockOptions.UserTags)
	assert.Equal(t, "value1", mockOptions.UserTags["tag1"])
	assert.Equal(t, "value2", mockOptions.UserTags["tag2"])

	// Test with invalid tags (odd number of arguments)
	opt2 := PutObjectTags("tag1", "value1", "tag2")
	err2 := opt2(mockOptions)
	assert.Error(t, err2)
	assert.Contains(t, err2.Error(), "received odd length input")
}

func TestPutObjectContentType(t *testing.T) {
	t.Parallel()

	contentType := "application/pdf"
	opt := PutObjectContentType(contentType)
	mockOptions := &minio.PutObjectOptions{}

	err := opt(mockOptions)
	assert.NoError(t, err)
	assert.Equal(t, contentType, mockOptions.ContentType)

	// Test with empty content type
	opt2 := PutObjectContentType("")
	err2 := opt2(mockOptions)
	assert.NoError(t, err2)
	assert.Equal(t, "", mockOptions.ContentType)
}

func TestDeleteObjectForced(t *testing.T) {
	t.Parallel()

	// Test with force = true
	opt := DeleteObjectForced(true)
	mockOptions := &minio.RemoveObjectOptions{}

	err := opt(mockOptions)
	assert.NoError(t, err)
	assert.True(t, mockOptions.ForceDelete)

	// Test with force = false
	opt2 := DeleteObjectForced(false)
	err2 := opt2(mockOptions)
	assert.NoError(t, err2)
	assert.False(t, mockOptions.ForceDelete)
}

func TestDeleteObjectBypassGovernancePolicy(t *testing.T) {
	t.Parallel()

	// Test with bypass = true
	opt := DeleteObjectBypassGovernancePolicy(true)
	mockOptions := &minio.RemoveObjectOptions{}

	err := opt(mockOptions)
	assert.NoError(t, err)
	assert.True(t, mockOptions.GovernanceBypass)

	// Test with bypass = false
	opt2 := DeleteObjectBypassGovernancePolicy(false)
	err2 := opt2(mockOptions)
	assert.NoError(t, err2)
	assert.False(t, mockOptions.GovernanceBypass)
}

func TestMinioConnectionPut(t *testing.T) {
	t.Parallel()

	// Test that PutObjectOptions processing works correctly with basic parameters
	mockOpts := &minio.PutObjectOptions{}

	// Test that all the functional options can be applied without panicking
	err1 := PutObjectMetadata("key", "value")(mockOpts)
	assert.NoError(t, err1)

	err2 := PutObjectTags("tag", "value")(mockOpts)
	assert.NoError(t, err2)

	err3 := PutObjectContentType("text/plain")(mockOpts)
	assert.NoError(t, err3)

	// Verify parameters were set correctly
	assert.Equal(t, "value", mockOpts.UserMetadata["key"])
	assert.Equal(t, "value", mockOpts.UserTags["tag"])
	assert.Equal(t, "text/plain", mockOpts.ContentType)
}

func TestMinioConnectionDelete(t *testing.T) {
	t.Parallel()

	// Test that DeleteObjectOptions processing works correctly with basic parameters
	mockOpts := &minio.RemoveObjectOptions{}

	// Test that all the functional options can be applied without panicking
	err1 := DeleteObjectForced(true)(mockOpts)
	assert.NoError(t, err1)

	err2 := DeleteObjectBypassGovernancePolicy(true)(mockOpts)
	assert.NoError(t, err2)

	// Verify parameters were set correctly
	assert.True(t, mockOpts.ForceDelete)
	assert.True(t, mockOpts.GovernanceBypass)
}
