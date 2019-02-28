package main

import (
	"fmt"
	"lc_xml_content/database"
	opencms_xml "lc_xml_content/xml"
	"strings"

	"github.com/PuerkitoBio/goquery"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

func main() {
	db, err := database.InitDatabase(`root@/ncfu?charset=utf8&parseTime=true&loc=Local`)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	importedXmls := opencms_xml.GetAllOpenCMSNews("./files-to-convert/")
	_, mapOfDatabaseLinks := database.GetAllLinksToResources(db)
	outputXMLs := make(map[string]*opencms_xml.OpenCMSNewsBlocks)

	for filePath, currentXML := range importedXmls {
		currentXML.XMLSchemaLocation = `opencms://system/modules/ru.soft.malikov.web/schemas/NewsBlock.xsd`
		currentXML.XMLAttr = `http://www.w3.org/2001/XMLSchema-instance`
		for linkIndex, currentLink := range currentXML.NewsBlock.FullDescription.Links.LinksList {
			validLink := strings.TrimSpace(currentLink.Target.Value)
			fmt.Printf("Slice size is: %d, current index is: %d\n", len(currentXML.NewsBlock.FullDescription.Links.LinksList), linkIndex)
			if mapOfDatabaseLinks[validLink] == "1" {

				currentXML.NewsBlock.FullDescription.Links.LinksList = removeFromLinksList(currentXML.NewsBlock.FullDescription.Links.LinksList, linkIndex)

				fmt.Printf("Link deleted in %s\n", filePath)
				fmt.Printf("Link index is: %d\n", linkIndex)
				content := currentXML.NewsBlock.FullDescription.Content
				strTofind := fmt.Sprintf(`a[href="%%(link%d)"]`, linkIndex)
				doc, _ := goquery.NewDocumentFromReader(strings.NewReader(content.Value))

				doc.Find(strTofind).Each(func(loop int, s *goquery.Selection) {
					fmt.Printf("Text of removed link: %s\n", s.Text())
					s.ReplaceWithSelection(s.Contents())
				})

				currentXML.NewsBlock.FullDescription.Content.Value, _ = doc.Html()
				outputXMLs[filePath] = currentXML
			} else {
				if val, flag := mapOfDatabaseLinks[validLink]; flag {
					currentLink.Target.Value = val
				}
				outputXMLs[filePath] = currentXML
				fmt.Printf("Link changed in %s\n link number is %d\n", filePath, linkIndex)
			}
		}
	}
	opencms_xml.ChangeXMLFiles(outputXMLs)
}

func removeFromLinksList(linksList []*opencms_xml.LinkToResource, linkNumber int) []*opencms_xml.LinkToResource {
	searchStr := fmt.Sprintf("link%d", linkNumber)
	for index, elem := range linksList {
		if elem.Name == searchStr {
			linksList = append(linksList[:index], linksList[index+1:]...)
		}
	}
	if len(linksList) > 0 {
		return linksList
	}
	return nil
}
