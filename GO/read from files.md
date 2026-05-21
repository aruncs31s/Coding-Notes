```go
func GetUniversityCodeByID(id string) string {
	jsonPath := "path.json"
	file, err := os.Open(jsonPath)
	if err != nil {
		return ""
	}
	defer file.Close()

	var colleges map[string]CollegeSettings
	err = json.NewDecoder(file).Decode(&colleges)
	if err != nil {
		return ""
	}
	settings, ok := colleges[id]
	if !ok {
		return ""
	}
	return settings.University
}

```