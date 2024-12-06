package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type jsonObj struct {
	Name string `json:"name"`
	Age  string `json:"age"`
}

type xmlObject struct {
	Name string `xml:"name"`
	Age  string `xml:"age"`
}

func createFile(baseFile string) error {
	create, err := os.Create(baseFile)
	defer create.Close()

	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	fmt.Println("Created file:", baseFile)
	return nil
}

func writeFile(filepath string, data string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("file %s not exists: %w", filepath, err)
	}
	defer file.Close()

	_, err = file.WriteString(data)
	if err != nil {
		return err
	}

	fmt.Println("Data were successfully written to", filepath)
	return nil
}

func readFile(filepath string) (string, error) {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	if strings.HasSuffix(filepath, ".json") {
		err = json.Indent(&buf, data, "", "  ")
		if err == nil {
			fmt.Println(buf.String())
		}
	} else {
		fmt.Printf("Data:\n %s\n", string(data))
	}
	return string(data), nil
}

func createJsonObject(filepath string) error {
	filepath = filepath + ".json"
	jsonStruct := jsonObj{Name: "Боб", Age: "20"}
	data, err := json.MarshalIndent(jsonStruct, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка при сериализации JSON: %w", err)
	}
	return writeFile(filepath, string(data))
}

func createXmlObject(filepath string) error {
	xmlStruct := xmlObject{Name: filepath, Age: "20"}
	data, err := xml.Marshal(xmlStruct)
	if err != nil {
		return fmt.Errorf("ошибка при сериализации xml: %w", err)
	}
	return writeFile(filepath, string(data))
}

func createZipArchive(zipPath, filePath string) error {
	zipPath = zipPath + "." + "zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании zip файла: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = addFileToZip(zipWriter, filePath)
	if err != nil {
		return err
	}

	fmt.Println("Файл успешно добавлен в zip архив:", zipPath)
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	fileToZip, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("не удалось открыть файл для добавления в zip: %w", err)
	}
	defer fileToZip.Close()

	w, err := zipWriter.Create(filepath.Base(fileToZip.Name()))
	if err != nil {
		return fmt.Errorf("не удалось создать запись в zip: %w", err)
	}

	_, err = io.Copy(w, fileToZip)
	if err != nil {
		return fmt.Errorf("не удалось записать файл в zip: %w", err)
	}

	return nil
}

func addFileToExistingZip(zipPath, newFilePath string) error {
	tempZipPath := zipPath + ".tmp"
	tempZip, err := os.Create(tempZipPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании временного zip файла: %w", err)
	}
	defer tempZip.Close()

	oldZip, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("ошибка при открытии zip файла: %w", err)
	}
	defer oldZip.Close()

	zipWriter := zip.NewWriter(tempZip)
	defer zipWriter.Close()

	for _, file := range oldZip.File {
		oldFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("ошибка при чтении файла из старого zip архива: %w", err)
		}
		defer oldFile.Close()

		newFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return fmt.Errorf("ошибка при создании файла в новом zip архиве: %w", err)
		}

		_, err = io.Copy(newFile, oldFile)
		if err != nil {
			return fmt.Errorf("ошибка при копировании данных в новый zip архив: %w", err)
		}
	}

	err = addFileToZip(zipWriter, newFilePath)
	if err != nil {
		return fmt.Errorf("ошибка при добавлении нового файла в zip архив: %w", err)
	}

	err = os.Rename(tempZipPath, zipPath)
	if err != nil {
		return fmt.Errorf("ошибка при замене старого архива новым: %w", err)
	}

	fmt.Println("Новый файл успешно добавлен в архив:", zipPath)
	return nil
}

func unzipArchive(zipPath, destDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("не удалось открыть zip файл: %w", err)
	}
	defer r.Close()

	for _, f := range r.File {
		destPath := filepath.Join(destDir, f.Name)
		fmt.Println("Распаковка файла:", destPath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
			continue
		}

		err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
		if err != nil {
			return fmt.Errorf("не удалось создать директорию для файла %s: %w", destPath, err)
		}

		outFile, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("не удалось создать файл при распаковке %s: %w", destPath, err)
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return fmt.Errorf("не удалось открыть файл %s в zip архиве: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return fmt.Errorf("ошибка копирования данных %s: %w", destPath, err)
		}
	}
	fmt.Println("Распаковка завершена:", destDir)
	return nil
}

func removeFileFromZip(zipPath, fileToRemove string) error {
	tempZipPath := zipPath + ".tmp"
	tempZip, err := os.Create(tempZipPath)
	if err != nil {
		return fmt.Errorf("ошибка при создании временного zip файла: %w", err)
	}
	defer tempZip.Close()

	// Открытие исходного архива
	oldZip, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("ошибка при открытии zip файла: %w", err)
	}
	defer oldZip.Close()

	zipWriter := zip.NewWriter(tempZip)
	defer zipWriter.Close()

	for _, file := range oldZip.File {
		if file.Name == fileToRemove {
			fmt.Println("Файл удален из архива:", fileToRemove)
			continue
		}

		oldFile, err := file.Open()
		if err != nil {
			return fmt.Errorf("ошибка при чтении файла из старого zip архива: %w", err)
		}
		defer oldFile.Close()

		newFile, err := zipWriter.Create(file.Name)
		if err != nil {
			return fmt.Errorf("ошибка при создании файла в новом zip архиве: %w", err)
		}

		_, err = io.Copy(newFile, oldFile)
		if err != nil {
			return fmt.Errorf("ошибка при копировании данных в новый zip архив: %w", err)
		}
	}

	err = os.Rename(tempZipPath, zipPath)
	if err != nil {
		return fmt.Errorf("ошибка при замене старого архива новым: %w", err)
	}

	return nil
}

func getDiskInfo() {
	cmd := exec.Command("df", "-Th")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Ошибка:", err)
		return
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines[1:] {
		if len(line) > 0 {
			fields := strings.Fields(line)
			if len(fields) >= 7 {
				fmt.Printf("Диск: %s\n", fields[0])                 // Имя диска
				fmt.Printf("Тип файловой системы: %s\n", fields[1]) // Тип файловой системы
				fmt.Printf("Размер: %s\n", fields[2])               // Общий размер
				fmt.Printf("Использовано: %s\n", fields[3])         // Использовано
				fmt.Printf("Свободно: %s\n", fields[4])             // Свободно
				fmt.Printf("Метка: %s\n\n", fields[5])              // Монтирование
			}
		}
	}
}

func createXFile(filepath, state, data string) error {
	filepath = filepath + "." + state
	err := createFile(filepath)
	if err != nil {
		return err
	}

	str, err := readFile(data)
	if err != nil {
		_ = writeFile(filepath, data)
	} else {
		_ = writeFile(filepath, str)
	}
	return nil
}

func main() {
	rootCmd := &cobra.Command{Use: "m"}

	rootCmd.AddCommand(&cobra.Command{
		Use:   "diskinfo",
		Short: "Получить информацию о дисках",
		Run: func(cmd *cobra.Command, args []string) {
			getDiskInfo()
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "createjsonfile [filepath] [data/file]",
		Short: "Создать json файл",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createXFile(args[0], "json", args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "createxmlfile [filepath] [data/file]",
		Short: "Создать json файл",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createXFile(args[0], "xml", args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "createfile [filepath]",
		Short: "Создать файл",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createFile(args[0]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	})

	writeCmd := &cobra.Command{
		Use:   "writefile [filepath] [data]",
		Short: "Записать данные в файл",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := writeFile(args[0], args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(writeCmd)

	readCmd := &cobra.Command{
		Use:   "readfile [filepath]",
		Short: "Прочитать данные из файла",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if _, err := readFile(args[0]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(readCmd)

	createJSONCmd := &cobra.Command{
		Use:   "createjsonobject [filepath]",
		Short: "Создать JSON объект",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createJsonObject(args[0]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(createJSONCmd)

	createXMLCmd := &cobra.Command{
		Use:   "createxmlobject [filepath]",
		Short: "Создать XML объект",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createXmlObject(args[0]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(createXMLCmd)

	createZipCmd := &cobra.Command{
		Use:   "createzip [zipPath] [filePath]",
		Short: "Создать ZIP архив",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := createZipArchive(args[0], args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(createZipCmd)

	addZipCmd := &cobra.Command{
		Use:   "addtozip [zipPath] [filePath]",
		Short: "Добавить файл в существующий ZIP архив",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := addFileToExistingZip(args[0], args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(addZipCmd)

	unzipCmd := &cobra.Command{
		Use:   "unzip [zipPath] [destDir]",
		Short: "Распаковать ZIP архив",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := unzipArchive(args[0], args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(unzipCmd)

	removeZipCmd := &cobra.Command{
		Use:   "removefromzip [zipPath] [fileToRemove]",
		Short: "Удалить файл из ZIP архива",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			if err := removeFileFromZip(args[0], args[1]); err != nil {
				fmt.Println("Error:", err)
			}
		},
	}
	rootCmd.AddCommand(removeZipCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
