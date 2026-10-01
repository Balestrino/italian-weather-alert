package config

// Storage secrets are loaded only for explicit storage operations. The public
// process reads them only when retained-copy access is explicitly enabled.
type Storage struct{ Bucket, AccessKey, SecretKey string }

func LoadStorage() (Storage, error) {
	c := Storage{Bucket: env("IWA_RUSTFS_BUCKET", "iwa-originals")}
	var err error
	c.AccessKey, err = secret("IWA_RUSTFS_ACCESS_KEY_FILE")
	if err != nil {
		return Storage{}, err
	}
	c.SecretKey, err = secret("IWA_RUSTFS_SECRET_KEY_FILE")
	if err != nil {
		return Storage{}, err
	}
	return c, nil
}
