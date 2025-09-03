package database

import (
	"testogo/internal/model/entity"
	"testogo/pkg/database"
)

// SeedData seeds initial data for the application
func SeedData() error {
	// Seed grades
	grades := []entity.Grade{
		{Name: "一年级", Code: "grade1", Description: "6-7岁", AgeMin: 6, AgeMax: 7, Order: 1, IsActive: true},
		{Name: "二年级", Code: "grade2", Description: "7-8岁", AgeMin: 7, AgeMax: 8, Order: 2, IsActive: true},
		{Name: "三年级", Code: "grade3", Description: "8-9岁", AgeMin: 8, AgeMax: 9, Order: 3, IsActive: true},
		{Name: "四年级", Code: "grade4", Description: "9-10岁", AgeMin: 9, AgeMax: 10, Order: 4, IsActive: true},
		{Name: "五年级", Code: "grade5", Description: "10-11岁", AgeMin: 10, AgeMax: 11, Order: 5, IsActive: true},
		{Name: "六年级", Code: "grade6", Description: "11-12岁", AgeMin: 11, AgeMax: 12, Order: 6, IsActive: true},
	}

	for _, grade := range grades {
		var existing entity.Grade
		if err := database.DB.Where("code = ?", grade.Code).First(&existing).Error; err != nil {
			database.DB.Create(&grade)
		}
	}

	// Seed subjects
	subjects := []entity.Subject{
		{
			Name:        "数学",
			Code:        "math",
			Description: "数学计算与逻辑思维",
			Icon:        "Money",
			Color:       "#667eea",
			Order:       1,
			IsActive:    true,
		},
		{
			Name:        "语言词汇",
			Code:        "vocabulary", 
			Description: "词汇积累与语言表达",
			Icon:        "ChatDotRound",
			Color:       "#f093fb",
			Order:       2,
			IsActive:    true,
		},
		{
			Name:        "阅读",
			Code:        "reading",
			Description: "阅读理解与文字感悟",
			Icon:        "Reading",
			Color:       "#4facfe",
			Order:       3,
			IsActive:    true,
		},
		{
			Name:        "识字",
			Code:        "literacy",
			Description: "汉字认识与书写练习",
			Icon:        "EditPen",
			Color:       "#43e97b",
			Order:       4,
			IsActive:    true,
		},
	}

	for _, subject := range subjects {
		var existing entity.Subject
		if err := database.DB.Where("code = ?", subject.Code).First(&existing).Error; err != nil {
			if err := database.DB.Create(&subject).Error; err != nil {
				return err
			}
		}
	}

	// Get subject IDs for topics
	var mathSubject entity.Subject
	database.DB.Where("code = ?", "math").First(&mathSubject)
	
	var vocabularySubject entity.Subject
	database.DB.Where("code = ?", "vocabulary").First(&vocabularySubject)
	
	var readingSubject entity.Subject
	database.DB.Where("code = ?", "reading").First(&readingSubject)
	
	var literacySubject entity.Subject
	database.DB.Where("code = ?", "literacy").First(&literacySubject)

	// Seed topics
	topics := []entity.Topic{
		// Math topics
		{
			SubjectID:       mathSubject.ID,
			Name:            "加法",
			Code:            "addition",
			Description:     "加法运算练习",
			FullDescription: "掌握基础加法运算，培养数学计算能力",
			Icon:            "Plus",
			Color:           "#667eea",
			Order:           1,
			IsActive:        true,
		},
		{
			SubjectID:       mathSubject.ID,
			Name:            "减法",
			Code:            "subtraction",
			Description:     "减法运算练习",
			FullDescription: "掌握基础减法运算，提升逻辑思维能力",
			Icon:            "Minus",
			Color:           "#f093fb",
			Order:           2,
			IsActive:        true,
		},
		{
			SubjectID:       mathSubject.ID,
			Name:            "乘法",
			Code:            "multiplication",
			Description:     "乘法运算练习",
			FullDescription: "掌握乘法口诀，提升计算速度和准确性",
			Icon:            "Star",
			Color:           "#fa709a",
			Order:           3,
			IsActive:        true,
		},
		{
			SubjectID:       mathSubject.ID,
			Name:            "除法",
			Code:            "division",
			Description:     "除法运算练习",
			FullDescription: "理解除法概念，掌握除法运算技巧",
			Icon:            "Close",
			Color:           "#a8edea",
			Order:           4,
			IsActive:        true,
		},
		{
			SubjectID:       mathSubject.ID,
			Name:            "应用题",
			Code:            "application",
			Description:     "数学应用练习",
			FullDescription: "通过实际生活情境，提升数学解决问题的能力",
			Icon:            "Document",
			Color:           "#4facfe",
			Order:           5,
			IsActive:        true,
		},
		// Vocabulary topics
		{
			SubjectID:       vocabularySubject.ID,
			Name:            "基础词汇",
			Code:            "word_basic",
			Description:     "常用词汇学习",
			FullDescription: "掌握日常生活中的基础词汇，提升语言表达能力",
			Icon:            "Star",
			Color:           "#f093fb",
			Order:           1,
			IsActive:        true,
		},
		{
			SubjectID:       vocabularySubject.ID,
			Name:            "近义词",
			Code:            "word_synonym",
			Description:     "近义词辨析",
			FullDescription: "学习词语的近义词，提升语言理解和运用能力",
			Icon:            "Document",
			Color:           "#43e97b",
			Order:           2,
			IsActive:        true,
		},
		// Reading topics
		{
			SubjectID:       readingSubject.ID,
			Name:            "故事阅读",
			Code:            "story_reading",
			Description:     "故事理解练习",
			FullDescription: "通过故事阅读，提升阅读理解和想象力",
			Icon:            "Reading",
			Color:           "#4facfe",
			Order:           1,
			IsActive:        true,
		},
		{
			SubjectID:       readingSubject.ID,
			Name:            "诗歌朗读",
			Code:            "poetry_reading",
			Description:     "诗歌欣赏练习",
			FullDescription: "通过诗歌朗读，培养语感和文学素养",
			Icon:            "Star",
			Color:           "#fa709a",
			Order:           2,
			IsActive:        true,
		},
		// Literacy topics
		{
			SubjectID:       literacySubject.ID,
			Name:            "汉字识别",
			Code:            "char_recognition",
			Description:     "汉字认识练习",
			FullDescription: "学习常用汉字，提升识字能力",
			Icon:            "EditPen",
			Color:           "#43e97b",
			Order:           1,
			IsActive:        true,
		},
		{
			SubjectID:       literacySubject.ID,
			Name:            "汉字书写",
			Code:            "char_writing",
			Description:     "汉字书写练习",
			FullDescription: "练习汉字正确书写，培养良好书写习惯",
			Icon:            "Document",
			Color:           "#667eea",
			Order:           2,
			IsActive:        true,
		},
	}

	for _, topic := range topics {
		var existing entity.Topic
		if err := database.DB.Where("subject_id = ? AND code = ?", topic.SubjectID, topic.Code).First(&existing).Error; err != nil {
			database.DB.Create(&topic)
		}
	}

	// Seed some basic reinforcement items
	reinforcementItems := []entity.ReinforcementItem{
		{
			Name:          "小红花",
			Type:          entity.ItemTypeFlower,
			MediaURL:      "",
			Color:         "#f56c6c",
			Icon:          "flower",
			Duration:      3000,
			AnimationType: "bounce",
			IsActive:      true,
		},
		{
			Name:          "金星星",
			Type:          entity.ItemTypeStar,
			MediaURL:      "",
			Color:         "#f0a957",
			Icon:          "star",
			Duration:      3000,
			AnimationType: "sparkle",
			IsActive:      true,
		},
		{
			Name:          "优秀徽章",
			Type:          entity.ItemTypeBadge,
			MediaURL:      "",
			Color:         "#67c23a",
			Icon:          "badge",
			Duration:      4000,
			AnimationType: "slide-down",
			IsActive:      true,
		},
		{
			Name:          "烟花动画",
			Type:          entity.ItemTypeFireworks,
			MediaURL:      "",
			Color:         "#409eff",
			Icon:          "fireworks",
			Duration:      5000,
			AnimationType: "fireworks",
			IsActive:      true,
		},
	}

	for _, item := range reinforcementItems {
		var existing entity.ReinforcementItem
		if err := database.DB.Where("name = ?", item.Name).First(&existing).Error; err != nil {
			database.DB.Create(&item)
		}
	}

	return nil
}