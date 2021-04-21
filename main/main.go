package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

const (
	UNIT_ROW         int = 0
	CLASS_YEAR_ROW   int = 1
	DISCIPLINE_ROW   int = 2
	HEADER_EXCEL_ROW int = 3
)

func isUnit(unitExpected string, unitInput string) bool {
	return strings.Contains(unitInput, unitExpected)
}

func main() {
	f, err := excelize.OpenFile("../resource/notas.xlsx")

	if err != nil {
		println(err.Error())
		return
	}

	sheetList := f.GetSheetList()
	classMap := make(map[string]map[string]map[string]map[string]string)
	var disciplines []string
	var students []string

	for _, sheet := range sheetList {
		rows, err := f.GetRows(sheet)
		if err != nil {
			println(err.Error())
			return
		}

		log.Printf("GetRows: %v", rows)

		discipline := ""
		class := ""
		unit := ""

		for index, row := range rows {
			if index == UNIT_ROW {
				unit = row[0]
				continue
			}

			if index == CLASS_YEAR_ROW {
				class = row[0]
				continue
			}

			if index == DISCIPLINE_ROW {
				discipline = row[0]
				continue
			}

			if index <= HEADER_EXCEL_ROW {
				continue
			}

			student := row[0]
			mean := row[5]

			if classMap[class] == nil {
				classMap[class] = map[string]map[string]map[string]string{}
				students = append(students, student)
			}

			if classMap[class][student] == nil {
				classMap[class][student] = map[string]map[string]string{}
				students = append(students, student)
			}

			if classMap[class][student][discipline] == nil {
				classMap[class][student][discipline] = map[string]string{}
			}
			log.Printf("Map indexes: %s, %s, %s, %s", student, discipline, strings.ToLower(unit), mean)

			classMap[class][student][discipline][unit] = mean
		}
	}

	log.Printf("Map Student: %+v", classMap)
	log.Printf("Map Disciplines: %+v", disciplines)

	for _, student := range students {
		for class, studentMap := range classMap {
			generatePdf(class, student, studentMap)
		}
	}
}

func getSchoolHeader() []string {
	return []string{
		"Educandário Ideal",
	}
}

func getClassHeader() []string {
	return []string{
		"Turma",
	}
}

func getStudentHeader() []string {
	return []string{
		"Nome do Aluno",
	}
}

func getTableHeader() []string {
	return []string{
		"Disciplina",
		"Média - 1º Unidade",
		"Média - 2º Unidade",
		"Média - 3º Unidade",
		"Média - 4º Unidade",
		"Média Final",
	}
}

func calculateFinalMean(means map[string]string) float64 {
	sum := func(means map[string]string) float64 {
		result := 0.0
		for _, v := range means {
			intVal, _ := strconv.ParseFloat(v, 64)
			result += intVal
		}
		return result
	}

	return sum(means) / 4
}

func formatContents(student string, studentMap map[string]map[string]map[string]string) [][]string {
	var studentMeans [][]string

	for discipline, _ := range studentMap[student] {

		disciplineData := []string{discipline, "-", "-", "-", "-", "-"}

		unitsMean := studentMap[student][discipline]

		for unit, mean := range unitsMean {
			for i := 1; i <= 4; i++ {
				if isUnit(fmt.Sprint(i), unit) {
					disciplineData[i] = mean
				}
			}
		}

		if len(unitsMean) == 4 {
			disciplineData[5] = fmt.Sprintf("%.2f", calculateFinalMean(unitsMean))
		}

		studentMeans = append(studentMeans, disciplineData)
	}

	return studentMeans
}

func generatePdf(class string, student string, studentMap map[string]map[string]map[string]string) {
	whiteColor := color.NewWhite()
	schoolHeader := getSchoolHeader()
	classHeader := getClassHeader()
	studentHeader := getStudentHeader()
	header := getTableHeader()

	contents := formatContents(student, studentMap)

	log.Printf("Contents: %+v", contents)

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 15, 10)
	m.SetBackgroundColor(whiteColor)

	m.Row(20, func() {
		m.Col(3, func() {
			_ = m.FileImage("../resource/image/educandario_ideal.jpg", props.Rect{
				Center:  true,
				Percent: 80,
			})
		})
		m.Col(9, func() {
			m.TableList(schoolHeader, [][]string{[]string{"Boletim de Notas"}}, props.TableList{
				HeaderProp: props.TableListContent{
					Size:      15,
					GridSizes: []uint{12},
				},
				ContentProp: props.TableListContent{
					Size:      10,
					GridSizes: []uint{12},
				},
				Align:              consts.Center,
				HeaderContentSpace: 1,
				Line:               false,
			})
		})
	})

	m.Row(5, func() {})

	m.TableList(classHeader, [][]string{[]string{class}}, props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{12},
		},
		ContentProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{12},
		},
		Align:              consts.Center,
		HeaderContentSpace: 1,
		Line:               false,
	})

	m.Row(5, func() {})

	m.TableList(studentHeader, [][]string{[]string{student}}, props.TableList{
		HeaderProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{12},
		},
		ContentProp: props.TableListContent{
			Size:      9,
			GridSizes: []uint{12},
		},
		Align:              consts.Center,
		HeaderContentSpace: 1,
		Line:               false,
	})

	m.Row(10, func() {})

	m.Row(20, func() {
		m.TableList(header, contents, props.TableList{
			HeaderProp: props.TableListContent{
				Size:      9,
				GridSizes: []uint{2, 2, 2, 2, 2, 2},
			},
			ContentProp: props.TableListContent{
				Size:      8,
				GridSizes: []uint{2, 2, 2, 2, 2, 2},
			},
			Align:              consts.Center,
			HeaderContentSpace: 1,
			Line:               true,
		})
	})
	err := m.OutputFileAndClose(fmt.Sprintf("../resource/%s-%s.pdf", student, class))
	if err != nil {
		log.Println("Could not save PDF:", err)
		os.Exit(1)
	}
}
