package cfgtypes

import "time"

type Duration time.Duration

func (d *Duration) UnmarshalText(b []byte) error {
	duration, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

func (d Duration) MarshalText() ([]byte, error) {
	return []byte(time.Duration(d).String()), nil
}

func (d Duration) AsDuration() time.Duration {
	return time.Duration(d)
}

func (d Duration) String() string {
	return time.Duration(d).String()
}
