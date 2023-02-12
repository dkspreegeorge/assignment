package PeriodTask

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/nav-inc/datetime"
)

const TIMEFORMAT = "20060102T150405Z"
const DSTFORMAT = "-0700"

// validate period here
func (task *periodicTask) validate() error {
	yearPeriod, monthPeriod, dayPeriod, hourPeriod, minutePeriod, secondsPeriod, err := getPeriodValues(task.Period)
	if err != nil {
		fmt.Println("error in getPeriodValues:", err)
		return fmt.Errorf(`[
"status": "error",
"desc": "Unsupported period"
]`)
	}

	log.Println("yearPeriod:", yearPeriod)
	log.Println("monthPeriod:", monthPeriod)
	log.Println("dayPeriod:", dayPeriod)
	log.Println("hourPeriod:", hourPeriod)
	log.Println("minutePeriod:", minutePeriod)
	log.Println("secondsPeriod:", secondsPeriod)

	return nil
}

// public method to use on main
func HandleGetRequest(w http.ResponseWriter, r *http.Request) {
	var task periodicTask

	// get the variable from query of the url
	query := r.URL.Query()

	task.Period = query.Get("period")
	task.Tz = query.Get("tz")

	tz, err := time.LoadLocation(query.Get("tz"))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"desc":   "Time zone is not valid.",
		})
		return
	}
	task.Tz = tz.String()

	//check if t1 is valid
	t1, err := datetime.Parse(query.Get("t1"), time.UTC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"desc":   "t1 value is invalid.",
		})
		return
	}
	task.T1 = t1

	//check if t2 is valid
	t2, err := datetime.Parse(query.Get("t2"), time.UTC)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "error",
			"desc":   "t2 value is invalid.",
		})
		return
	}
	task.T2 = t2

	//check if period is right
	if err := task.validate(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(err.Error()))
		return
	}

	//check if t2 is after t1
	if err := task.validateTime(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(err.Error()))
		return
	}

	//here create the response

	//here pass it to a json to return it
	jsonTask, _ := json.MarshalIndent(findAllTimestamps(task), "", "")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonTask)
}

// simple struct for period
type periodicTask struct {
	Period string    `json:"period"`
	Tz     string    `json:"tz"`
	T1     time.Time `json:"t1"`
	T2     time.Time `json:"t2"`
}

// check if t2 is earlier than t1
func (task *periodicTask) validateTime() error {
	if task.T2.Before(task.T1) {
		return fmt.Errorf(`{"status": "error", "desc": "t2 can not be earlier than t1."}`)
	}
	return nil
}

func findAllTimestamps(task periodicTask) []string {
	startTime := task.T1
	endTime := task.T2
	period := task.Period
	startTime = startTime.UTC()
	//======================================================Get the period from file cause we need it dynamically
	yearPeriod, monthPeriod, dayPeriod, hourPeriod, minutePeriod, secondsPeriod, err := getPeriodValues(period)
	if err != nil {
		log.Println("Error reading file:", err)

	}
	var timestamps []string
	log.Println("Starting to find the timestamps")
	log.Println("The start time is " + startTime.String())

	//=====================format time according to specifications========month only and year only add one more hour
	if monthPeriod != 0 || yearPeriod != 0 {
		startTime = startTime.Add(time.Hour)
	}
	//move to next hour
	duration := time.Duration((60-startTime.Minute()))*time.Minute + time.Duration((60-startTime.Second()))*time.Second
	startTime = startTime.Add(duration).Truncate(time.Hour)
	offset1 := startTime.Local().Format(DSTFORMAT)
	log.Println("The start time is changed to " + startTime.String())
	for startTime.Before(endTime) {
		// check if DCT
		offset2 := startTime.Local().Format(DSTFORMAT)
		startTime = adjustTime(startTime, offset1, offset2)

		//month only being handled in a special way
		if monthPeriod != 0 && yearPeriod == 0 && dayPeriod == 0 && hourPeriod == 0 && minutePeriod == 0 && secondsPeriod == 0 {
			lastDay := time.Date(startTime.Year(), startTime.Month()+1, 0, 0, 0, 0, 0, startTime.Location())
			lastDay = time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), lastDay.Location())

			// check if DCT
			offset2 = startTime.Local().Format(DSTFORMAT)

			if startTime.Day() == lastDay.Day() {
				// check if DCT
				offset2 = startTime.Local().Format(DSTFORMAT)

				startTime = adjustTime(startTime, offset1, offset2)
				if endTime.Before(startTime) {
					break
				}
				log.Println("1Adding timestamp of :" + startTime.Format(TIMEFORMAT))
				timestamps = append(timestamps, startTime.Format(TIMEFORMAT))
				offset1 = startTime.Local().Format(DSTFORMAT)

				//just need to add one day here
				startTime = startTime.AddDate(0, 0, 1)
			} else {

				startTime = lastDay
				// check if DCT
				offset2 := startTime.Local().Format(DSTFORMAT)
				startTime = adjustTime(startTime, offset1, offset2)
				if endTime.Before(startTime) {
					break
				}
				log.Println("2Adding timestamp of :" + startTime.Format(TIMEFORMAT))
				timestamps = append(timestamps, startTime.Format(TIMEFORMAT))
				offset1 = startTime.Local().Format(DSTFORMAT)

				//just need to add one day here
				startTime = startTime.AddDate(0, 0, 1)

			}

			//year being handled in a special way
		} else if yearPeriod != 0 && monthPeriod == 0 && dayPeriod == 0 && hourPeriod == 0 && minutePeriod == 0 && secondsPeriod == 0 {
			lastDay := time.Date(startTime.Year(), 12, 31, 0, 0, 0, 0, startTime.Location())
			startTime = time.Date(lastDay.Year(), lastDay.Month(), lastDay.Day(), startTime.Hour(), startTime.Minute(), startTime.Second(), startTime.Nanosecond(), lastDay.Location())

			offset2 = startTime.Local().Format(DSTFORMAT)

			offset2 = startTime.Local().Format(DSTFORMAT)

			startTime = adjustTime(startTime, offset1, offset2)
			if endTime.Before(startTime) {
				break
			}
			log.Println("3Adding timestamp of :" + startTime.Format(TIMEFORMAT))
			timestamps = append(timestamps, startTime.Format(TIMEFORMAT))
			offset1 = startTime.Local().Format(DSTFORMAT)
			startTime = startTime.AddDate(1, 0, 0)

		} else {
			log.Println("4Adding timestamp of :" + startTime.Format(TIMEFORMAT))
			timestamps = append(timestamps, startTime.Format(TIMEFORMAT))
			offset1 = startTime.Local().Format(DSTFORMAT)
			startTime = startTime.AddDate(yearPeriod, monthPeriod, dayPeriod) //add days months or years
			startTime = startTime.Add(time.Duration(minutePeriod) * time.Minute)
			startTime = startTime.Add(time.Duration(hourPeriod) * time.Hour)
			startTime = startTime.Add(time.Duration(secondsPeriod) * time.Second)

		}

	}

	return timestamps
}

// method to get values from the json config file
func getPeriodValues(period string) (yearPeriod, monthPeriod, dayPeriod, hourPeriod, minutePeriod, secondsPeriod int, err error) {
	path, err := filepath.Abs("./configPeriod.json")
	if err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Failed to get absolute path: %v", err)
	}

	file, err := ioutil.ReadFile(path)
	if err != nil {
		path, err = filepath.Abs("./api/configPeriod.json")
		file, err = ioutil.ReadFile(path)
		if err != nil {
			return 0, 0, 0, 0, 0, 0, fmt.Errorf("Failed to read file: %v", err)
		}
	}

	var periodValues map[string][]int
	if err := json.Unmarshal(file, &periodValues); err != nil {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("Failed to unmarshal JSON: %v", err)
	}

	value, ok := periodValues[period]
	if !ok {
		return 0, 0, 0, 0, 0, 0, fmt.Errorf("The key does not exist in the map")
	}

	return value[0], value[1], value[2], value[3], value[4], value[5], nil
}

// method to adjust time daylight changes
func adjustTime(startTime time.Time, preoffset1, preoffset2 string) time.Time {
	if preoffset1 != preoffset2 {
		if preoffset1 > preoffset2 {
			log.Println("Time changed due to daylight changes.Current is: " + startTime.String())
			startTime = startTime.Add(time.Hour)
			log.Println("New time:" + startTime.String())
		} else {
			log.Println("Time changed due to daylight changes.Current is: " + startTime.String())
			startTime = startTime.Add(-time.Hour)
			log.Println("New time:" + startTime.String())
		}
	}
	return startTime
}
