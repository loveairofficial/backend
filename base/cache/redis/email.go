package redis

import "time"

func (r *Redis) SetPin(key string, pin string, d time.Duration) error {
	ctx, cancel := getContext()
	defer cancel()

	err := r.localClient0.Set(ctx, key, pin, d).Err()
	return err
}

func (r *Redis) GetPin(key string) (string, error) {
	ctx, cancel := getContext()
	defer cancel()

	res, err := r.localClient0.Get(ctx, key).Result()

	return res, err
}

func (r *Redis) DeletePin(key string) error {
	ctx, cancel := getContext()
	defer cancel()

	err := r.localClient0.Del(ctx, key).Err()

	return err
}
