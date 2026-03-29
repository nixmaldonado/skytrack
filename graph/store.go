package graph

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/nixmaldonado/skytrack/graph/model"
)

type AirportStore struct {
	mu       sync.RWMutex
	airports []model.Airport
	nextID   int
}

func NewAirportStore() *AirportStore {
	iata := func(s string) *model.IATACode { c := model.IATACode(s); return &c }
	city := func(s string) *string { return &s }
	elev := func(i int) *int { return &i }

	store := &AirportStore{
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
		nextID: 11,
	}
	return store
}

func (s *AirportStore) All() []model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Airport, len(s.airports))
	copy(result, s.airports)
	return result
}

func (s *AirportStore) FindByICAO(icao model.ICAOCode) *model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.airports {
		if a.Icao == icao {
			cp := a
			return &cp
		}
	}
	return nil
}

func (s *AirportStore) FindByIATA(iata model.IATACode) *model.Airport {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, a := range s.airports {
		if a.Iata != nil && *a.Iata == iata {
			cp := a
			return &cp
		}
	}
	return nil
}

func (s *AirportStore) Create(input model.CreateAirportInput) (*model.Airport, error) {
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

func (s *AirportStore) Update(icao model.ICAOCode, input model.UpdateAirportInput) (*model.Airport, error) {
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
