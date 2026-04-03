package opensky

// StateVector represents a single aircraft state from the OpenSky Network API.
// Pointer fields are nullable — OpenSky returns null when data is unavailable
// (e.g., aircraft on the ground with transponder off).
type StateVector struct {
	Icao24       string
	Callsign     string
	Longitude    *float64
	Latitude     *float64
	BaroAltitude *float64
	Velocity     *float64
	TrueTrack    *float64 // heading in degrees clockwise from north
	VerticalRate *float64
	OnGround     bool
	TimePosition *int64
}
