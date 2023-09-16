package losses

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	s "strings"
	"time"
)

type LDate time.Time

type StatisticOfLoses struct {
	Message string `json:"message"`
	Data    struct {
		Date     LDate     `json:"date"`
		Resource string    `json:"resource"`
		Status   WarStatus `json:"war_status"`
		Stats    Stat      `json:"stats"`
		Increase Stat      `json:"increase"`
		Day      int       `json:"day"`
	} `json:"data"`
}

type WarStatus struct {
	Alias string `json:"alias"`
	Code  int    `json:"code"`
}

type Stat struct {
	PersonnelUnits           int `json:"personnel_units"`
	Tanks                    int `json:"tanks"`
	ArmouredFightingVehicles int `json:"armoured_fighting_vehicles"`
	ArtillerySystems         int `json:"artillery_systems"`
	Mlrs                     int `json:"mlrs"`
	AaWarfareSystems         int `json:"aa_warfare_systems"`
	Planes                   int `json:"planes"`
	Helicopters              int `json:"helicopters"`
	VehiclesFuelTanks        int `json:"vehicles_fuel_tanks"`
	WarshipsCutters          int `json:"warships_cutters"`
	CruiseMissiles           int `json:"cruise_missiles"`
	UavSystems               int `json:"uav_systems"`
	SpecialMilitaryEquip     int `json:"special_military_equip"`
	AtgmSrbmSystems          int `json:"atgm_srbm_systems"`
	Submarines               int `json:"submarines"`
}

func (i StatisticOfLoses) ToMessage() string {
	d := i.Data
	stats := d.Stats
	incr := d.Increase
	msg := `
Втрати окупантів станом на %s (*%d день* повномаштабного вторгення росії в Україну) склали:
Особового складу: %d (*+%d*)
Танки: %d (*+%d*)
Бойові броньовані машини: %d (*+%d*)
Артилерійські системи: %d (*+%d*)
Реактивні системи залпового вогню: %d (*+%d*)
Засоби протиповітряної оборони: %d (*+%d*)
Літаки: %d (*+%d*)
Гелікоптери: %d (*+%d*)
Автомобільна техніка та цистерни з паливом: %d (*+%d*)
Кораблі/Катери: %d (*+%d*)
Крилаті ракети: %d (*+%d*)
Безпілотні літальні апарати: %d (*+%d*)
Спецтехніка: %d (*+%d*)
Установки ОТРК/ТРК: %d (*+%d*)
Підводні човни: %d (*+%d*)

Посилання на зведення генштабу: %s
`
	return fmt.Sprintf(
		msg,
		d.Date.String(), d.Day,
		stats.PersonnelUnits, incr.PersonnelUnits,
		stats.Tanks, incr.Tanks,
		stats.ArmouredFightingVehicles, incr.ArmouredFightingVehicles,
		stats.ArtillerySystems, incr.ArtillerySystems,
		stats.Mlrs, incr.Mlrs,
		stats.AaWarfareSystems, incr.AaWarfareSystems,
		stats.Planes, incr.Planes,
		stats.Helicopters, incr.Helicopters,
		stats.VehiclesFuelTanks, incr.VehiclesFuelTanks,
		stats.WarshipsCutters, incr.WarshipsCutters,
		stats.CruiseMissiles, incr.CruiseMissiles,
		stats.UavSystems, incr.UavSystems,
		stats.SpecialMilitaryEquip, incr.SpecialMilitaryEquip,
		stats.AtgmSrbmSystems, incr.AtgmSrbmSystems,
		stats.Submarines, incr.Submarines,
		d.Resource,
	)
}

func GetFreshInfo() (*StatisticOfLoses, error) {
	resp, err := http.Get("https://russianwarship.rip/api/v2/statistics/latest")
	if err != nil {
		return nil, err
	}

	respBytes, readError := io.ReadAll(resp.Body)
	if readError != nil {
		return nil, readError
	}

	var info StatisticOfLoses
	parseError := json.Unmarshal(respBytes, &info)
	if parseError != nil {
		return nil, parseError
	}

	return &info, nil
}

func (d *LDate) UnmarshalJSON(b []byte) error {
	value := s.Trim(string(b), `"`) // get rid of "
	if value == "" || value == "null" {
		return nil
	}

	t, err := time.Parse("2006-01-02", value) // parse time
	if err != nil {
		return err
	}

	*d = LDate(t) // set result using the pointer
	return nil
}

func (d *LDate) String() string {
	return time.Time(*d).Format("02.01.2006")
}

type FindImgError struct {
	msg string
}

func (e FindImgError) Error() string {
	return e.msg
}
