package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type docRefEntry struct {
	XMLName    xml.Name `xml:"refentry"`
	Id         string   `xml:"id,attr"`
	RefName    string   `xml:"refnamediv>refname"`
	RefPurpose string   `xml:"refnamediv>refpurpose"`
}

type docSvnIndex struct {
	XMLName xml.Name     `xml:"svn"`
	FileRef []docFileRef `xml:"index>file"`
}

type docFileRef struct {
	Ref string `xml:"href,attr"`
}

type docFile struct {
	BaseName string
	FileName string
}

type CommandDoc struct {
	BaseName string
	Purpose  string
	// TODO: more?
}

type CommandDocs struct {
	MajorVersion int
	Commands     []*CommandDoc
}

type Documentation struct {
	CommandDocs []CommandDocs
}

func makeCmdDocUrl(cmdName string, majorVersion int) string {
	manVer := "2"
	if majorVersion >= 3 {
		manVer = strconv.Itoa(majorVersion)
	}
	return fmt.Sprintf("https://www.opengl.org/sdk/docs/man%s/xhtml/gl%s.xml", manVer, cmdName)
}

func makeGLDocUrl(majorVersion int) string {
	manVer := "2"
	if majorVersion >= 3 {
		manVer = strconv.Itoa(majorVersion)
	}
	return fmt.Sprintf("https://www.opengl.org/sdk/docs/man%s", manVer)
}

func makeExtenionSpecDocUrl(vendor, extension string) string {
	return fmt.Sprintf("https://www.opengl.org/registry/specs/%s/%s.txt", vendor, extension)
}

func (cd CommandDocs) Len() int           { return len(cd.Commands) }
func (cd CommandDocs) Swap(i, j int)      { cd.Commands[i], cd.Commands[j] = cd.Commands[j], cd.Commands[i] }
func (cd CommandDocs) Less(i, j int) bool { return cd.Commands[i].BaseName < cd.Commands[j].BaseName }

func (d *Documentation) findCmd(majorVersion int, cmdName string) (*CommandDoc, error) {
	if majorVersion == 1 {
		majorVersion = 2
	}
	for _, cd := range d.CommandDocs {
		if cd.MajorVersion == majorVersion {
			index := sort.Search(len(cd.Commands), func(i int) bool {
				return cd.Commands[i].BaseName >= cmdName
			})
			if index == len(cd.Commands) {
				return nil, fmt.Errorf("Command doc not found: %s, version %d", cmdName, majorVersion)
			}
			if strings.HasPrefix(cmdName, cd.Commands[index].BaseName) {
				return cd.Commands[index], nil
			}
			if index == 0 {
				return nil, fmt.Errorf("Command doc not found: %s, version %d", cmdName, majorVersion)
			}
			if strings.HasPrefix(cmdName, cd.Commands[index-1].BaseName) {
				return cd.Commands[index-1], nil
			}
			return nil, fmt.Errorf("Command doc not found: %s, version %d", cmdName, majorVersion)
		}
	}
	return nil, fmt.Errorf("Version not found %d", majorVersion)
}

func (d *Documentation) AnnotatePackages(ps Packages) {
	for _, p := range ps {
		for _, f := range p.Functions {
			cd, err := d.findCmd(p.Version.Major, f.Name)
			if err == nil {
				f.Doc = fmt.Sprintf("%s (%s)", cd.Purpose, makeCmdDocUrl(cd.BaseName, p.Version.Major))
			}
		}
	}
}

func readXmlFileNonStrict(fileName string, data interface{}) error {
	reader, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer reader.Close()
	decoder := xml.NewDecoder(reader)
	decoder.Strict = false
	return decoder.Decode(data)
}

func parseDocIndex(fileName string) ([]docFile, error) {
	var di docSvnIndex
	if err := readXmlFileNonStrict(fileName, &di); err != nil {
		return nil, err
	}
	files := make([]docFile, 0, 256)
	for _, fr := range di.FileRef {
		if strings.HasPrefix(fr.Ref, "glu") { // ignore
			continue
		}
		if strings.HasPrefix(fr.Ref, "glX") { // ignore
			continue
		}
		if strings.HasPrefix(fr.Ref, "gl") {
			fn := strings.TrimPrefix(strings.TrimSuffix(fr.Ref, ".xml"), "gl")
			files = append(files, docFile{BaseName: fn, FileName: fr.Ref})
		}
	}
	return files, nil
}

func DownloadDocs(url, docCat, outDir string) error {
	complOutDir := filepath.Join(outDir, docCat)
	err := downloadFile(url, docCat, complOutDir, "index.xml")
	if err != nil {
		return err
	}
	file, err := parseDocIndex(filepath.Join(complOutDir, "index.xml"))
	if err != nil {
		return err
	}
	for _, file := range file {
		err = downloadFile(fmt.Sprintf("%s/%s", url, docCat), file.FileName, complOutDir, file.FileName)
		if err != nil {
			return err
		}
	}
	return nil
}

func parseDocFile(fileName string) (*CommandDoc, error) {
	var d docRefEntry
	if err := readXmlFileNonStrict(fileName, &d); err != nil {
		return nil, err
	}
	return &CommandDoc{Purpose: d.RefPurpose}, nil
}

func parseDocs(docCat, dir string) ([]*CommandDoc, error) {
	complOutDir := filepath.Join(dir, docCat)
	file, err := parseDocIndex(filepath.Join(complOutDir, "index.xml"))
	if err != nil {
		return nil, err
	}
	commandDocs := make([]*CommandDoc, 0, 256)
	for _, file := range file {
		cd, err := parseDocFile(filepath.Join(complOutDir, file.FileName))
		if err != nil {
			return nil, err
		}
		cd.BaseName = file.BaseName
		commandDocs = append(commandDocs, cd)
	}
	return commandDocs, nil
}

func ParseAllDocs(dir string) (*Documentation, error) {
	cdocs := make([]CommandDocs, 0, 4)
	for ver := 2; ver <= 4; ver++ {
		cds, err := parseDocs(fmt.Sprintf("man%d", ver), dir)
		if err != nil {
			return nil, err
		}
		cd := CommandDocs{MajorVersion: ver, Commands: cds}
		cdocs = append(cdocs, cd)
	}
	return &Documentation{CommandDocs: cdocs}, nil
}
