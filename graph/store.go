package graph

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/nixmaldonado/skytrack/graph/model"
)

type Store struct {
	mu       sync.RWMutex
	airports []model.Airport
	airlines []model.Airline
	flights  []flight
	nextID   int
	log      func(string) // logs "database queries" to demonstrate N+1
}

// flight is internal — stores IDs for relationships instead of resolved objects.
type flight struct {
	ID                 string
	Callsign           string
	AirlineID          string
	DepartureAirportID string
	ArrivalAirportID   string
	Status             model.FlightStatus
}

func NewStore() *Store {
	iata := func(s string) *model.IATACode { c := model.IATACode(s); return &c }
	city := func(s string) *string { return &s }
	elev := func(i int) *int { return &i }
	str := func(s string) *string { return &s }

	store := &Store{
		airports: []model.Airport{
			{ID: "1", Icao: "KJFK", Iata: iata("JFK"), Name: "John F Kennedy International Airport", City: city("New York"), Country: "US", Latitude: 40.6398, Longitude: -73.7789, Elevation: elev(13), Type: model.AirportTypeLarge},
			{ID: "2", Icao: "EGLL", Iata: iata("LHR"), Name: "Heathrow Airport", City: city("London"), Country: "GB", Latitude: 51.4706, Longitude: -0.4619, Elevation: elev(83), Type: model.AirportTypeLarge},
			{ID: "3", Icao: "RJTT", Iata: iata("HND"), Name: "Tokyo Haneda Airport", City: city("Tokyo"), Country: "JP", Latitude: 35.5523, Longitude: 139.7798, Elevation: elev(35), Type: model.AirportTypeLarge},
			{ID: "4", Icao: "LFPG", Iata: iata("CDG"), Name: "Charles de Gaulle Airport", City: city("Paris"), Country: "FR", Latitude: 49.0097, Longitude: 2.5479, Elevation: elev(392), Type: model.AirportTypeLarge},
			{ID: "5", Icao: "KLAX", Iata: iata("LAX"), Name: "Los Angeles International Airport", City: city("Los Angeles"), Country: "US", Latitude: 33.9425, Longitude: -118.4081, Elevation: elev(126), Type: model.AirportTypeLarge},
			{ID: "6", Icao: "OMDB", Iata: iata("DXB"), Name: "Dubai International Airport", City: city("Dubai"), Country: "AE", Latitude: 25.2528, Longitude: 55.3644, Elevation: elev(62), Type: model.AirportTypeLarge},
			{ID: "7", Icao: "VHHH", Iata: iata("HKG"), Name: "Hong Kong International Airport", City: city("Hong Kong"), Country: "HK", Latitude: 22.3089, Longitude: 113.9145, Elevation: elev(28), Type: model.AirportTypeLarge},
			{ID: "8", Icao: "WSSS", Iata: iata("SIN"), Name: "Singapore Changi Airport", City: city("Singapore"), Country: "SG", Latitude: 1.3502, Longitude: 103.9944, Elevation: elev(22), Type: model.AirportTypeLarge},
			{ID: "9", Icao: "EDDF", Iata: iata("FRA"), Name: "Frankfurt Airport", City: city("Frankfurt"), Country: "DE", Latitude: 50.0333, Longitude: 8.5706, Elevation: elev(364), Type: model.AirportTypeLarge},
			{ID: "10", Icao: "LEMD", Iata: iata("MAD"), Name: "Adolfo Suárez Madrid–Barajas Airport", City: city("Madrid"), Country: "ES", Latitude: 40.4719, Longitude: -3.5626, Elevation: elev(2000), Type: model.AirportTypeLarge},
		},
		airlines: []model.Airline{
			{ID: "1", Icao: "AAL", Iata: str("AA"), Name: "American Airlines", Country: "US"},
			{ID: "2", Icao: "BAW", Iata: str("BA"), Name: "British Airways", Country: "GB"},
			{ID: "3", Icao: "UAE", Iata: str("EK"), Name: "Emirates", Country: "AE"},
			{ID: "4", Icao: "DLH", Iata: str("LH"), Name: "Lufthansa", Country: "DE"},
			{ID: "5", Icao: "AFR", Iata: str("AF"), Name: "Air France", Country: "FR"},
		},
		flights: []flight{
			{ID: "1", Callsign: "AAL100", AirlineID: "1", DepartureAirportID: "1", ArrivalAirportID: "2", Status: model.FlightStatusActive},
			{ID: "2", Callsign: "BAW178", AirlineID: "2", DepartureAirportID: "2", ArrivalAirportID: "1", Status: model.FlightStatusActive},
			{ID: "3", Callsign: "UAE201", AirlineID: "3", DepartureAirportID: "6", ArrivalAirportID: "3", Status: model.FlightStatusScheduled},
			{ID: "4", Callsign: "DLH456", AirlineID: "4", DepartureAirportID: "9", ArrivalAirportID: "5", Status: model.FlightStatusActive},
			{ID: "5", Callsign: "AFR662", AirlineID: "5", DepartureAirportID: "4", ArrivalAirportID: "7", Status: model.FlightStatusLanded},
			{ID: "6", Callsign: "AAL455", AirlineID: "1", DepartureAirportID: "5", ArrivalAirportID: "8", Status: model.FlightStatusScheduled},
			{ID: "7", Callsign: "BAW015", AirlineID: "2", DepartureAirportID: "2", ArrivalAirportID: "8", Status: model.FlightStatusActive},
			{ID: "8", Callsign: "UAE773", AirlineID: "3", DepartureAirportID: "6", ArrivalAirportID: "10", Status: model.FlightStatusActive},
		},
		nextID: 11,
		log:    func(msg string) { fmt.Println("📊 " + msg) },
	}
	return store
}

// ── Airport methods ──

func (s *Store) All() []model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log("SELECT * FROM airports")
	result := make([]model.Airport, len(s.airports))
	copy(result, s.airports)
	return result
}

func (s *Store) FindByICAO(icao model.ICAOCode) *model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log(fmt.Sprintf("SELECT * FROM airports WHERE icao = '%s'", icao))
	for _, a := range s.airports {
		if a.Icao == icao {
			cp := a
			return &cp
		}
	}
	return nil
}

func (s *Store) FindByIATA(iata model.IATACode) *model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log(fmt.Sprintf("SELECT * FROM airports WHERE iata = '%s'", iata))
	for _, a := range s.airports {
		if a.Iata != nil && *a.Iata == iata {
			cp := a
			return &cp
		}
	}
	return nil
}

func (s *Store) FindAirportByID(id string) *model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log(fmt.Sprintf("SELECT * FROM airports WHERE id = '%s'", id))
	for _, a := range s.airports {
		if a.ID == id {
			cp := a
			return &cp
		}
	}
	return nil
}

func (s *Store) Create(input model.CreateAirportInput) (*model.Airport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, a := range s.airports {
		if a.Icao == input.Icao {
			return nil, fmt.Errorf("airport with ICAO code %q already exists", input.Icao)
		}
	}

	airport := model.Airport{
		ID:        strconv.Itoa(s.nextID),
		Icao:      input.Icao,
		Iata:      input.Iata,
		Name:      input.Name,
		City:      input.City,
		Country:   input.Country,
		Latitude:  input.Latitude,
		Longitude: input.Longitude,
		Elevation: input.Elevation,
		Type:      input.Type,
	}
	s.nextID++
	s.airports = append(s.airports, airport)
	return &airport, nil
}

func (s *Store) Update(icao model.ICAOCode, input model.UpdateAirportInput) (*model.Airport, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, a := range s.airports {
		if a.Icao == icao {
			if input.Iata != nil {
				s.airports[i].Iata = input.Iata
			}
			if input.Name != nil {
				s.airports[i].Name = *input.Name
			}
			if input.City != nil {
				s.airports[i].City = input.City
			}
			if input.Country != nil {
				s.airports[i].Country = *input.Country
			}
			if input.Latitude != nil {
				s.airports[i].Latitude = *input.Latitude
			}
			if input.Longitude != nil {
				s.airports[i].Longitude = *input.Longitude
			}
			if input.Elevation != nil {
				s.airports[i].Elevation = input.Elevation
			}
			if input.Type != nil {
				s.airports[i].Type = *input.Type
			}
			cp := s.airports[i]
			return &cp, nil
		}
	}
	return nil, fmt.Errorf("airport with ICAO code %q not found", icao)
}

func (s *Store) Search(query string) []model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	q := strings.ToLower(query)
	var results []model.Airport
	for _, a := range s.airports {
		if strings.Contains(strings.ToLower(a.Name), q) {
			results = append(results, a)
			continue
		}

		if a.City != nil && strings.Contains(strings.ToLower(*a.City), q) {
			results = append(results, a)
		}
	}

	return results
}

// ── Airline methods ──

func (s *Store) FindAirlineByID(id string) *model.Airline {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log(fmt.Sprintf("SELECT * FROM airlines WHERE id = '%s'", id))
	for _, a := range s.airlines {
		if a.ID == id {
			cp := a
			return &cp
		}
	}
	return nil
}

// ── Flight methods ──

func (s *Store) AllFlights() []model.Flight {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log("SELECT * FROM flights")
	result := make([]model.Flight, len(s.flights))
	for i, f := range s.flights {
		result[i] = model.Flight{
			ID:       f.ID,
			Callsign: f.Callsign,
			Status:   f.Status,
		}
	}
	return result
}

// FlightAirlineID returns the airline ID for a flight (simulates a FK lookup).
func (s *Store) FlightAirlineID(flightID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.flights {
		if f.ID == flightID {
			return f.AirlineID
		}
	}
	return ""
}

// FlightDepartureAirportID returns the departure airport ID for a flight.
func (s *Store) FlightDepartureAirportID(flightID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.flights {
		if f.ID == flightID {
			return f.DepartureAirportID
		}
	}
	return ""
}

// FlightArrivalAirportID returns the arrival airport ID for a flight.
func (s *Store) FlightArrivalAirportID(flightID string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, f := range s.flights {
		if f.ID == flightID {
			return f.ArrivalAirportID
		}
	}
	return ""
}
