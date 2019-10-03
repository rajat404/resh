package records

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/mattn/go-shellwords"
)

// BaseRecord - common base for Record and FallbackRecord
type BaseRecord struct {
	// core
	CmdLine   string `json:"cmdLine"`
	ExitCode  int    `json:"exitCode"`
	Shell     string `json:"shell"`
	Uname     string `json:"uname"`
	SessionID string `json:"sessionId"`

	// posix
	Home  string `json:"home"`
	Lang  string `json:"lang"`
	LcAll string `json:"lcAll"`
	Login string `json:"login"`
	//Path     string `json:"path"`
	Pwd      string `json:"pwd"`
	PwdAfter string `json:"pwdAfter"`
	ShellEnv string `json:"shellEnv"`
	Term     string `json:"term"`

	// non-posix"`
	RealPwd      string `json:"realPwd"`
	RealPwdAfter string `json:"realPwdAfter"`
	Pid          int    `json:"pid"`
	SessionPid   int    `json:"sessionPid"`
	Host         string `json:"host"`
	Hosttype     string `json:"hosttype"`
	Ostype       string `json:"ostype"`
	Machtype     string `json:"machtype"`
	Shlvl        int    `json:"shlvl"`

	// before after
	TimezoneBefore string `json:"timezoneBefore"`
	TimezoneAfter  string `json:"timezoneAfter"`

	RealtimeBefore      float64 `json:"realtimeBefore"`
	RealtimeAfter       float64 `json:"realtimeAfter"`
	RealtimeBeforeLocal float64 `json:"realtimeBeforeLocal"`
	RealtimeAfterLocal  float64 `json:"realtimeAfterLocal"`

	RealtimeDuration          float64 `json:"realtimeDuration"`
	RealtimeSinceSessionStart float64 `json:"realtimeSinceSessionStart"`
	RealtimeSinceBoot         float64 `json:"realtimeSinceBoot"`
	//Logs []string      `json: "logs"`

	GitDir          string `json:"gitDir"`
	GitRealDir      string `json:"gitRealDir"`
	GitOriginRemote string `json:"gitOriginRemote"`
	MachineID       string `json:"machineId"`

	OsReleaseID         string `json:"osReleaseId"`
	OsReleaseVersionID  string `json:"osReleaseVersionId"`
	OsReleaseIDLike     string `json:"osReleaseIdLike"`
	OsReleaseName       string `json:"osReleaseName"`
	OsReleasePrettyName string `json:"osReleasePrettyName"`

	ReshUUID     string `json:"reshUuid"`
	ReshVersion  string `json:"reshVersion"`
	ReshRevision string `json:"reshRevision"`

	// added by sanitizatizer
	Sanitized bool `json:"sanitized,omitempty"`
	CmdLength int  `json:"cmdLength,omitempty"`
}

// Record representing single executed command with its metadata
type Record struct {
	BaseRecord

	Cols  string `json:"cols"`
	Lines string `json:"lines"`
}

// EnrichedRecord - record enriched with additional data
type EnrichedRecord struct {
	Record

	// enriching fields - added "later"
	Command             string   `json:"command"`
	FirstWord           string   `json:"firstWord"`
	Invalid             bool     `json:"invalid"`
	SeqSessionID        uint64   `json:"seqSessionId"`
	LastRecordOfSession bool     `json:"lastRecordOfSession"`
	DebugThisRecord     bool     `json:"debugThisRecord"`
	Errors              []string `json:"errors"`
	// SeqSessionID uint64 `json:"seqSessionId,omitempty"`
}

// FallbackRecord when record is too old and can't be parsed into regular Record
type FallbackRecord struct {
	BaseRecord
	// older version of the record where cols and lines are int

	Cols  int `json:"cols"`  // notice the int type
	Lines int `json:"lines"` // notice the int type
}

// ConvertRecord from FallbackRecord to Record
func ConvertRecord(r *FallbackRecord) Record {
	return Record{
		BaseRecord: r.BaseRecord,
		// these two lines are the only reason we are doing this
		Cols:  strconv.Itoa(r.Cols),
		Lines: strconv.Itoa(r.Lines),
	}
}

// ToString - returns record the json
func (r EnrichedRecord) ToString() (string, error) {
	jsonRec, err := json.Marshal(r)
	if err != nil {
		return "marshalling error", err
	}
	return string(jsonRec), nil
}

// Enriched - returnd enriched record
func Enriched(r Record) EnrichedRecord {
	record := EnrichedRecord{Record: r}
	// Get command/first word from commandline
	var err error
	record.Command, record.FirstWord, err = GetCommandAndFirstWord(r.CmdLine)
	if err != nil {
		record.Errors = append(record.Errors, "GetCommandAndFirstWord error:"+err.Error())
		rec, _ := record.ToString()
		log.Println("Invalid command:", rec)
		record.Invalid = true
		return record
	}
	err = r.Validate()
	if err != nil {
		record.Errors = append(record.Errors, "Validate error:"+err.Error())
		rec, _ := record.ToString()
		log.Println("Invalid command:", rec)
		record.Invalid = true
	}
	return record
	// TODO: Detect and mark simple commands r.Simple
}

// Validate - returns error if the record is invalid
func (r *Record) Validate() error {
	if r.CmdLine == "" {
		return errors.New("There is no CmdLine")
	}
	if r.RealtimeBefore == 0 || r.RealtimeAfter == 0 {
		return errors.New("There is no Time")
	}
	if r.RealtimeBeforeLocal == 0 || r.RealtimeAfterLocal == 0 {
		return errors.New("There is no Local Time")
	}
	if r.RealPwd == "" || r.RealPwdAfter == "" {
		return errors.New("There is no Real Pwd")
	}
	if r.Pwd == "" || r.PwdAfter == "" {
		return errors.New("There is no Pwd")
	}

	// TimezoneBefore
	// TimezoneAfter

	// RealtimeDuration
	// RealtimeSinceSessionStart - TODO: add later
	// RealtimeSinceBoot  - TODO: add later

	// device extras
	// Host
	// Hosttype
	// Ostype
	// Machtype
	// OsReleaseID
	// OsReleaseVersionID
	// OsReleaseIDLike
	// OsReleaseName
	// OsReleasePrettyName

	// session extras
	// Term
	// Shlvl

	// static info
	// Lang
	// LcAll

	// meta
	// ReshUUID
	// ReshVersion
	// ReshRevision

	// added by sanitizatizer
	// Sanitized
	// CmdLength
	return nil
}

// SetCmdLine sets cmdLine and related members
func (r *EnrichedRecord) SetCmdLine(cmdLine string) {
	r.CmdLine = cmdLine
	r.CmdLength = len(cmdLine)
	r.ExitCode = 0
	var err error
	r.Command, r.FirstWord, err = GetCommandAndFirstWord(cmdLine)
	if err != nil {
		r.Errors = append(r.Errors, "GetCommandAndFirstWord error:"+err.Error())
		// log.Println("Invalid command:", r.CmdLine)
		r.Invalid = true
	}
}

// Stripped returns record stripped of all info that is not available during prediction
func Stripped(r EnrichedRecord) EnrichedRecord {
	// clear the cmd itself
	r.SetCmdLine("")
	// replace after info with before info
	r.PwdAfter = r.Pwd
	r.RealPwdAfter = r.RealPwd
	r.TimezoneAfter = r.TimezoneBefore
	r.RealtimeAfter = r.RealtimeBefore
	r.RealtimeAfterLocal = r.RealtimeBeforeLocal
	// clear some more stuff
	r.RealtimeDuration = 0
	r.LastRecordOfSession = false
	return r
}

// GetCommandAndFirstWord func
func GetCommandAndFirstWord(cmdLine string) (string, string, error) {
	args, err := shellwords.Parse(cmdLine)
	if err != nil {
		log.Println("shellwords Error:", err, " (cmdLine: <", cmdLine, "> )")
		return "", "", err
	}
	if len(args) == 0 {
		return "", "", nil
	}
	i := 0
	for true {
		// commands in shell sometimes look like this `variable=something command argument otherArgument --option`
		//		to get the command we skip over tokens that contain '='
		if strings.ContainsRune(args[i], '=') && len(args) > i+1 {
			i++
			continue
		}
		return args[i], args[0], nil
	}
	log.Fatal("GetCommandAndFirstWord error: this should not happen!")
	return "ERROR", "ERROR", errors.New("this should not happen - contact developer ;)")
}

// DistParams is used to supply params to Enrichedrecords.DistanceTo()
type DistParams struct {
	ExitCode  float64
	MachineID float64
	SessionID float64
	Login     float64
	Shell     float64
	Pwd       float64
	RealPwd   float64
	Git       float64
	Time      float64
}

// DistanceTo another record
func (r *EnrichedRecord) DistanceTo(r2 EnrichedRecord, p DistParams) float64 {
	var dist float64
	dist = 0

	// lev distance or something? TODO later
	// CmdLine

	// exit code
	if r.ExitCode != r2.ExitCode {
		if r.ExitCode == 0 || r2.ExitCode == 0 {
			// one success + one error -> 1
			dist += 1 * p.ExitCode
		} else {
			// two different errors
			dist += 0.5 * p.ExitCode
		}
	}

	// machine/device
	if r.MachineID != r2.MachineID {
		dist += 1 * p.MachineID
	}
	// Uname

	// session
	if r.SessionID != r2.SessionID {
		dist += 1 * p.SessionID
	}
	// Pid - add because of nested shells?
	// SessionPid

	// user
	if r.Login != r2.Login {
		dist += 1 * p.Login
	}
	// Home

	// shell
	if r.Shell != r2.Shell {
		dist += 1 * p.Shell
	}
	// ShellEnv

	// pwd
	if r.Pwd != r2.Pwd {
		// TODO: compare using hierarchy
		// TODO: make more important
		dist += 1 * p.Pwd
	}
	if r.RealPwd != r2.RealPwd {
		// TODO: -||-
		dist += 1 * p.RealPwd
	}
	// PwdAfter
	// RealPwdAfter

	// git
	if r.GitDir != r2.GitDir {
		dist += 1 * p.Git
	}
	if r.GitRealDir != r2.GitRealDir {
		dist += 1 * p.Git
	}
	if r.GitOriginRemote != r2.GitOriginRemote {
		dist += 1 * p.Git
	}

	// time
	// this can actually get negative for differences of less than one second which is fine
	// distance grows by 1 with every order
	distTime := math.Log10(math.Abs(r.RealtimeBefore-r2.RealtimeBefore)) * p.Time
	if math.IsNaN(distTime) == false && math.IsInf(distTime, 0) == false {
		dist += distTime
	}
	// RealtimeBeforeLocal
	// RealtimeAfter
	// RealtimeAfterLocal

	// TimezoneBefore
	// TimezoneAfter

	// RealtimeDuration
	// RealtimeSinceSessionStart - TODO: add later
	// RealtimeSinceBoot  - TODO: add later

	// device extras
	// Host
	// Hosttype
	// Ostype
	// Machtype
	// OsReleaseID
	// OsReleaseVersionID
	// OsReleaseIDLike
	// OsReleaseName
	// OsReleasePrettyName

	// session extras
	// Term
	// Shlvl

	// static info
	// Lang
	// LcAll

	// meta
	// ReshUUID
	// ReshVersion
	// ReshRevision

	// added by sanitizatizer
	// Sanitized
	// CmdLength

	return dist
}
