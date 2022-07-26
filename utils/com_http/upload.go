package com_http

import (
	"errors"
	"io"
	"mime/multipart"
)

// UploadForm is the interface for http upload.
type UploadForm interface {
	// Write writes fields to multipart writer
	Write(w *multipart.Writer) error
}

// FormFileFunc writes file content to multipart writer.
type FormFileFunc func(w io.Writer) error

type formFile struct {
	fieldName string
	fileName  string
	fileFunc  FormFileFunc
}

type uploadForm struct {
	formFiles  []*formFile
	formFields map[string]string
}

func (f *uploadForm) Write(w *multipart.Writer) error {
	if len(f.formFiles) == 0 {
		return errors.New("empty file field")
	}

	for _, v := range f.formFiles {
		part, err := w.CreateFormFile(v.fieldName, v.fileName)

		if err != nil {
			return err
		}

		if err = v.fileFunc(part); err != nil {
			return err
		}
	}

	for name, value := range f.formFields {
		if err := w.WriteField(name, value); err != nil {
			return err
		}
	}

	return nil
}

// UploadField configures how we set up the upload from.
type UploadField func(f *uploadForm)

// WithFormFile specifies the file field to upload from.
func WithFormFile(fieldName, fileName string, fn FormFileFunc) UploadField {
	return func(f *uploadForm) {
		f.formFiles = append(f.formFiles, &formFile{
			fieldName: fieldName,
			fileName:  fileName,
			fileFunc:  fn,
		})
	}
}

// WithFormField specifies the form field to upload from.
func WithFormField(fieldName, fieldValue string) UploadField {
	return func(u *uploadForm) {
		u.formFields[fieldName] = fieldValue
	}
}

// NewUploadForm returns an upload form
func NewUploadForm(fields ...UploadField) UploadForm {
	form := &uploadForm{
		formFiles:  make([]*formFile, 0),
		formFields: make(map[string]string),
	}

	for _, f := range fields {
		f(form)
	}

	return form
}
