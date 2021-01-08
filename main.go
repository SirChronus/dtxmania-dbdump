package main

import (
	"bufio"
	"encoding/binary"
	"encoding/xml"
	"io"
	"log"
	"math"
	"os"
	"time"
)

type eType int32

const (
	DTX eType = iota
	GDA
	G2D
	BMS
	BME
	SMF
)

func (e eType) String() string {
	return [...]string{"DTX", "GDA", "G2D", "BMS", "BME", "SMF"}[e]
}

func (e eType) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	return enc.EncodeElement(e.String(), start)
}

type dateAsString string

type fileInformation struct {
	AbsoluteFilePath   string       `xml:"absolute-file-path"`
	AbsoluteFolderPath string       `xml:"absolute-folder-path"`
	LastModified       dateAsString `xml:"last-modified"`
	FileSize           int64        `xml:"file-size"`
}

type songIniInformation struct {
	LastModified dateAsString `xml:"last-modified"`
	FileSize     int64        `xml:"file-size"`
}

type dgbInt32 struct {
	Drums  int32 `xml:"drums"`
	Guitar int32 `xml:"guitar"`
	Bass   int32 `xml:"bass"`
}

type dgbDouble struct {
	Drums  float64 `xml:"drums"`
	Guitar float64 `xml:"guitar"`
	Bass   float64 `xml:"bass"`
}

type dgbBoolean struct {
	Drums  bool `xml:"drums"`
	Guitar bool `xml:"guitar"`
	Bass   bool `xml:"bass"`
}

type performanceHistory struct {
	First  string `xml:"first"`
	Second string `xml:"second"`
	Third  string `xml:"third"`
	Fourth string `xml:"fourth"`
	Fifth  string `xml:"fifth"`
}

type songInformation struct {
	Title              string             `xml:"title"`
	Artist             string             `xml:"artist"`
	Comment            string             `xml:"comment"`
	Genre              string             `xml:"genre"`
	PreImage           string             `xml:"pre-image"`
	PreMovie           string             `xml:"pre-movie"`
	PreSound           string             `xml:"pre-sound"`
	Background         string             `xml:"background"`
	Level              dgbInt32           `xml:"level"`
	LevelDec           dgbInt32           `xml:"level-dec"`
	BestRank           dgbInt32           `xml:"best-rank"`
	HighSkill          dgbDouble          `xml:"high-skill"`
	FullCombo          dgbBoolean         `xml:"full-combo"`
	NbPerformance      dgbInt32           `xml:"nb-performance"`
	PerformanceHistory performanceHistory `xml:"performance-history"`
	HiddenLevel        bool               `xml:"hidden-level"`
	Classic            dgbBoolean         `xml:"classic"`
	ScoreExists        dgbBoolean         `xml:"score-exists"`
	SongType           eType              `xml:"song-type"`
	Bpm                float64            `xml:"bpm"`
	Duration           int32              `xml:"duration"`
}

type score struct {
	XMLName            xml.Name           `xml:"song"`
	FileInformation    fileInformation    `xml:"file-info"`
	SongIniInformation songIniInformation `xml:"song-ini-info"`
	SongInformation    songInformation    `xml:"song-info"`
}

var fileReader *bufio.Reader
var file *os.File
var outFile *os.File

const tickFactor = 10000000

var isEOF = false

func logFatalIfError(err error) {
	if err != nil {
		if err == io.EOF {
			isEOF = true
			return
		}

		if file != nil {
			file.Close()
		}

		if outFile != nil {
			outFile.Close()
		}
		log.Fatalln(err)
	}
}

func readStringFromDBOrFail() string {
	length, err := binary.ReadUvarint(fileReader)
	logFatalIfError(err)

	stringAsBytes := make([]byte, length)
	_, err = io.ReadFull(fileReader, stringAsBytes)
	logFatalIfError(err)

	return string(stringAsBytes)
}

func readSignedInt64FromDBOrFail() int64 {
	valueAsBytes := make([]byte, 8)
	_, err := io.ReadFull(fileReader, valueAsBytes)
	logFatalIfError(err)

	return int64(binary.LittleEndian.Uint64(valueAsBytes))
}

func readSignedInt32FromDBOrFail() int32 {
	valueAsBytes := make([]byte, 4)
	_, err := io.ReadFull(fileReader, valueAsBytes)
	logFatalIfError(err)

	return int32(binary.LittleEndian.Uint32(valueAsBytes))
}

func readDoubleFromDBOrFail() float64 {
	valueAsBytes := make([]byte, 8)
	_, err := io.ReadFull(fileReader, valueAsBytes)
	logFatalIfError(err)

	return math.Float64frombits(binary.LittleEndian.Uint64(valueAsBytes))
}

func readBoolFromDBOrFail() bool {
	valueAsBytes := make([]byte, 1)
	_, err := io.ReadFull(fileReader, valueAsBytes)
	logFatalIfError(err)

	return valueAsBytes[0] != 0
}

func readDateFromDBOrFail() dateAsString {
	dateTime := readSignedInt64FromDBOrFail()
	baseTime := time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC).Unix()
	// Convert from C# tick time to proper UTC timestamp
	t := time.Unix(dateTime/tickFactor+baseTime, dateTime%tickFactor)

	return dateAsString(t.Format(time.RFC3339))
}

func readFileInformation(s *score) {
	s.FileInformation.AbsoluteFilePath = readStringFromDBOrFail()
	s.FileInformation.AbsoluteFolderPath = readStringFromDBOrFail()
	s.FileInformation.LastModified = readDateFromDBOrFail()
	s.FileInformation.FileSize = readSignedInt64FromDBOrFail()
}

func readSongIniInformation(s *score) {
	s.SongIniInformation.LastModified = readDateFromDBOrFail()
	s.SongIniInformation.FileSize = readSignedInt64FromDBOrFail()
}

func readDGBInt32(s *dgbInt32) {
	s.Drums = readSignedInt32FromDBOrFail()
	s.Guitar = readSignedInt32FromDBOrFail()
	s.Bass = readSignedInt32FromDBOrFail()
}

func readDGBDouble(s *dgbDouble) {
	s.Drums = readDoubleFromDBOrFail()
	s.Guitar = readDoubleFromDBOrFail()
	s.Bass = readDoubleFromDBOrFail()
}

func readDGBBoolean(s *dgbBoolean) {
	s.Drums = readBoolFromDBOrFail()
	s.Guitar = readBoolFromDBOrFail()
	s.Bass = readBoolFromDBOrFail()
}

func readPerfomanceHistory(s *performanceHistory) {
	s.First = readStringFromDBOrFail()
	s.Second = readStringFromDBOrFail()
	s.Third = readStringFromDBOrFail()
	s.Fourth = readStringFromDBOrFail()
	s.Fifth = readStringFromDBOrFail()
}

func readSongInformation(s *score) {
	s.SongInformation.Title = readStringFromDBOrFail()
	s.SongInformation.Artist = readStringFromDBOrFail()
	s.SongInformation.Comment = readStringFromDBOrFail()
	s.SongInformation.Genre = readStringFromDBOrFail()
	s.SongInformation.PreImage = readStringFromDBOrFail()
	s.SongInformation.PreMovie = readStringFromDBOrFail()
	s.SongInformation.PreSound = readStringFromDBOrFail()
	s.SongInformation.Background = readStringFromDBOrFail()
	readDGBInt32(&s.SongInformation.Level)
	readDGBInt32(&s.SongInformation.LevelDec)
	readDGBInt32(&s.SongInformation.BestRank)
	readDGBDouble(&s.SongInformation.HighSkill)
	readDGBBoolean(&s.SongInformation.FullCombo)
	readDGBInt32(&s.SongInformation.NbPerformance)
	readPerfomanceHistory(&s.SongInformation.PerformanceHistory)
	s.SongInformation.HiddenLevel = readBoolFromDBOrFail()
	readDGBBoolean(&s.SongInformation.Classic)
	readDGBBoolean(&s.SongInformation.ScoreExists)
	s.SongInformation.SongType = eType(readSignedInt32FromDBOrFail())
	s.SongInformation.Bpm = readDoubleFromDBOrFail()
	s.SongInformation.Duration = readSignedInt32FromDBOrFail()

}

func readScore(s *score) {
	readFileInformation(s)
	readSongIniInformation(s)
	readSongInformation(s)
}

func main() {
	file, err := os.Open("songs.db")
	logFatalIfError(err)
	defer file.Close()
	fileReader = bufio.NewReader(file)

	outFile, err = os.Create("dump.xml")
	logFatalIfError(err)
	defer outFile.Close()
	outFileWriter := bufio.NewWriter(outFile)
	_, err = outFileWriter.WriteString("<songs>\n")
	logFatalIfError(err)
	enc := xml.NewEncoder(outFileWriter)
	enc.Indent("  ", "    ")

	versionString := readStringFromDBOrFail()

	log.Printf("SongDB version: %s\n", versionString)
	for !isEOF {
		var s score
		readScore(&s)
		logFatalIfError(enc.Encode(s))
	}

	_, err = outFileWriter.WriteString("\n</songs>")
	logFatalIfError(outFileWriter.Flush())

	log.Println("done")
}
