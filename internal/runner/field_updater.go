package runner

import (
	"github.com/zhex/local-bbp/internal/common"
	"github.com/zhex/local-bbp/internal/models"
	"regexp"
	"strings"
)

type FieldUpdater struct {
	Secrets map[string]string
}

func NewFieldUpdater(secrets map[string]string) *FieldUpdater {
	return &FieldUpdater{
		Secrets: secrets,
	}
}

func (f *FieldUpdater) UpdateImage(image *models.Image) *models.Image {
	if image == nil {
		return nil
	}
	var newImage *models.Image
	_ = common.DeepClone(image, &newImage)

	f.Update(&newImage.Name)
	f.Update(&newImage.Username)
	f.Update(&newImage.Password)
	if newImage.AWS != nil {
		f.Update(&newImage.AWS.AccessKey)
		f.Update(&newImage.AWS.SecretKey)
		f.Update(&newImage.AWS.OIDCRole)
	}

	return newImage
}

func (f *FieldUpdater) UpdateMap(data map[string]string) map[string]string {
	if data == nil {
		return nil
	}
	newData := make(map[string]string)
	for k, v := range data {
		f.Update(&v)
		newData[k] = v
	}
	return newData
}

func (f *FieldUpdater) Update(field *string) {
	if field == nil {
		return
	}
	re := regexp.MustCompile(`\${(\w+)}`)
	*field = re.ReplaceAllStringFunc(*field, func(match string) string {
		key := match[2 : len(match)-1] // remove bracket
		if val, ok := f.Secrets[key]; ok {
			return val
		}
		return ""
	})
	if strings.HasPrefix(*field, "$") {
		key := (*field)[1:]
		if val, ok := f.Secrets[key]; ok {
			*field = val
		} else {
			*field = ""
		}
	}
}
