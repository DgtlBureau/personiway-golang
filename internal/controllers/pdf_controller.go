package controllers

import (
	"DgtlBureau/personiway-golang/internal/services"
	"log"
	"mime/multipart"
)

func RunConvert(fileHeaders []*multipart.FileHeader) {
	for _, fileHeader := range(fileHeaders){
		go func (){
			file, err := fileHeader.Open()

			if err != nil{
				log.Printf("Error opening file %s: %v", fileHeader.Filename, err)
				return
			}

			defer file.Close()

			services.Convert(file)
		}()

	}
}