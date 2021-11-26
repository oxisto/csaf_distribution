package csaf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// TLPLabel is the traffic light policy of the CSAF.
type TLPLabel string

const (
	// TLPLabelUnlabeled is the 'UNLABELED' policy.
	TLPLabelUnlabeled = "UNLABELED"
	// TLPLabelWhite is the 'WHITE' policy.
	TLPLabelWhite = "WHITE"
	// TLPLabelGreen is the 'GREEN' policy.
	TLPLabelGreen = "GREEN"
	// TLPLabelAmber is the 'AMBER' policy.
	TLPLabelAmber = "AMBER"
	// TLPLabelRed is the 'RED' policy.
	TLPLabelRed = "RED"
)

var tlpLabelPattern = alternativesUnmarshal(
	string("UNLABELED"),
	string("WHITE"),
	string("GREEN"),
	string("AMBER"),
	string("RED"))

// JSONURL is an URL to JSON document.
type JSONURL string

var jsonURLPattern = patternUnmarshal(`\.json$`)

// Feed is CSAF feed.
type Feed struct {
	Summary  string    `json:"summary"`
	TLPLabel *TLPLabel `json:"tlp_label"` // required
	URL      *JSONURL  `json:"url"`       // required
}

// ROLIE is the ROLIE extension of the CSAF feed.
type ROLIE struct {
	Categories []JSONURL `json:"categories,omitempty"`
	Feeds      []Feed    `json:"feeds"` // required
	Services   []JSONURL `json:"services,omitempty"`
}

// Distribution is a distribution of a CSAF feed.
type Distribution struct {
	DirectoryURL string  `json:"directory_url,omitempty"`
	Rolie        []ROLIE `json:"rolie"`
}

// TimeStamp represents a time stamp in a CSAF feed.
type TimeStamp time.Time

// Fingerprint is the fingerprint of a OpenPGP key used to sign
// the CASF documents.
type Fingerprint string

var fingerprintPattern = patternUnmarshal(`^[0-9a-fA-F]{40,}$`)

// PGPKey is location and the fingerprint of the key
// used to sign the CASF documents.
type PGPKey struct {
	Fingerprint Fingerprint `json:"fingerprint,omitempty"`
	URL         *string     `json:"url"` // required
}

// Category is the category of the CSAF feed.
type Category string

const (
	// CSAFCategoryCoordinator is the "coordinator" category.
	CSAFCategoryCoordinator Category = "coordinator"
	// CSAFCategoryDiscoverer is the "discoverer" category.
	CSAFCategoryDiscoverer Category = "discoverer"
	// CSAFCategoryOther is the "other" category.
	CSAFCategoryOther Category = "other"
	// CSAFCategoryTranslator is the "translator" category.
	CSAFCategoryTranslator Category = "translator"
	// CSAFCategoryUser is the "user" category.
	CSAFCategoryUser Category = "user"
	// CSAFCategoryVendor is the "vendor" category.
	CSAFCategoryVendor Category = "vendor"
)

var csafCategoryPattern = alternativesUnmarshal(
	string(CSAFCategoryCoordinator),
	string(CSAFCategoryDiscoverer),
	string(CSAFCategoryOther),
	string(CSAFCategoryTranslator),
	string(CSAFCategoryUser),
	string(CSAFCategoryVendor))

// Publisher is the publisher of the feed.
type Publisher struct {
	Category         *Category `json:"category"`  // required
	Name             *string   `json:"name"`      // required
	Namespace        *string   `json:"namespace"` // required
	ContactDetails   string    `json:"contact_details,omitempty"`
	IssuingAuthority string    `json:"issuing_authority,omitempty"`
}

// MetadataVersion is the metadata version of the feed.
type MetadataVersion string

// MetadataVersion20 is the current version of the schema.
const MetadataVersion20 MetadataVersion = "2.0"

var metadataVersionPattern = alternativesUnmarshal(string(MetadataVersion20))

// MetadataRole is the role of the feed.
type MetadataRole string

const (
	// MetadataRolePublisher is the "csaf_publisher" role.
	MetadataRolePublisher MetadataRole = "csaf_publisher"
	// MetadataRoleProvider is the "csaf_provider" role.
	MetadataRoleProvider MetadataRole = "csaf_provider"
	// MetadataRoleTrustedProvider is the "csaf_trusted_provider" role.
	MetadataRoleTrustedProvider MetadataRole = "csaf_trusted_provider"
)

var metadataRolePattern = alternativesUnmarshal(
	string(MetadataRolePublisher),
	string(MetadataRoleProvider),
	string(MetadataRoleTrustedProvider))

// ProviderURL is the URL of the provider document.
type ProviderURL string

var providerURLPattern = patternUnmarshal(`/provider-metadata\.json$`)

// ProviderMetadata contains the metadata of the provider.
type ProviderMetadata struct {
	CanonicalURL            *ProviderURL     `json:"canonical_url"` // required
	Distributions           []Distribution   `json:"distributions,omitempty"`
	LastUpdated             *TimeStamp       `json:"last_updated"` // required
	ListOnCSAFAggregators   *bool            `json:"list_on_CSAF_aggregators"`
	MetadataVersion         *MetadataVersion `json:"metadata_version"`           // required
	MirrorOnCSAFAggregators *bool            `json:"mirror_on_CSAF_aggregators"` // required
	PGPKeys                 []PGPKey         `json:"pgp_keys,omitempty"`
	Publisher               *Publisher       `json:"publisher"` // required
	Role                    *MetadataRole    `json:"role"`      // required
}

func patternUnmarshal(pattern string) func([]byte) (string, error) {
	r := regexp.MustCompile(pattern)
	return func(data []byte) (string, error) {
		s := string(data)
		if !r.MatchString(s) {
			return "", fmt.Errorf("%s does not match %v", s, r)
		}
		return s, nil
	}
}

func alternativesUnmarshal(alternatives ...string) func([]byte) (string, error) {
	return func(data []byte) (string, error) {
		s := string(data)
		for _, alt := range alternatives {
			if alt == s {
				return s, nil
			}
		}
		return "", fmt.Errorf("%s not in [%s]", s, strings.Join(alternatives, "|"))
	}
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (tl *TLPLabel) UnmarshalText(data []byte) error {
	s, err := tlpLabelPattern(data)
	if err == nil {
		*tl = TLPLabel(s)
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (ju *JSONURL) UnmarshalText(data []byte) error {
	s, err := jsonURLPattern(data)
	if err == nil {
		*ju = JSONURL(s)
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (pu *ProviderURL) UnmarshalText(data []byte) error {
	s, err := providerURLPattern(data)
	if err == nil {
		*pu = ProviderURL(s)
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (cc *Category) UnmarshalText(data []byte) error {
	s, err := csafCategoryPattern(data)
	if err == nil {
		*cc = Category(s)
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (fp *Fingerprint) UnmarshalText(data []byte) error {
	s, err := fingerprintPattern(data)
	if err == nil {
		*fp = Fingerprint(s)
	}
	return err
}

// UnmarshalText implements the encoding.TextUnmarshaller interface.
func (ts *TimeStamp) UnmarshalText(data []byte) error {
	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return err
	}
	*ts = TimeStamp(t)
	return nil
}

// MarshalText implements the encoding.TextMarshaller interface.
func (ts TimeStamp) MarshalText() ([]byte, error) {
	return []byte(time.Time(ts).Format(time.RFC3339)), nil
}

// Defaults fills the correct default values into the provider metadata.
func (pmd *ProviderMetadata) Defaults() {
	if pmd.Role == nil {
		role := MetadataRoleProvider
		pmd.Role = &role
	}
	if pmd.ListOnCSAFAggregators == nil {
		t := true
		pmd.ListOnCSAFAggregators = &t
	}
	if pmd.MirrorOnCSAFAggregators == nil {
		t := true
		pmd.MirrorOnCSAFAggregators = &t
	}
	if pmd.MetadataVersion == nil {
		mdv := MetadataVersion20
		pmd.MetadataVersion = &mdv
	}
}

// Validate checks if the feed is valid.
// Returns an error if the validation fails otherwise nil.
func (f *Feed) Validate() error {
	switch {
	case f.TLPLabel == nil:
		return errors.New("feed[].tlp_label is mandatory")
	case f.URL == nil:
		return errors.New("feed[].url is mandatory")
	}
	return nil
}

// Validate checks if the ROLIE extension is valid.
// Returns an error if the validation fails otherwise nil.
func (r *ROLIE) Validate() error {
	if len(r.Feeds) < 1 {
		return errors.New("ROLIE needs at least one feed")
	}
	for i := range r.Feeds {
		if err := r.Feeds[i].Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate checks if the publisher is valid.
// Returns an error if the validation fails otherwise nil.
func (cp *Publisher) Validate() error {
	switch {
	case cp == nil:
		return errors.New("publisher is mandatory")
	case cp.Category == nil:
		return errors.New("publisher.category is mandatory")
	case cp.Name == nil:
		return errors.New("publisher.name is mandatory")
	case cp.Namespace == nil:
		return errors.New("publisher.namespace is mandatory")
	}
	return nil
}

// Validate checks if the PGPKey is valid.
// Returns an error if the validation fails otherwise nil.
func (pk *PGPKey) Validate() error {
	if pk.URL == nil {
		return errors.New("pgp_key[].url is mandatory")
	}
	return nil
}

// Validate checks if the distribution is valid.
// Returns an error if the validation fails otherwise nil.
func (d *Distribution) Validate() error {
	for i := range d.Rolie {
		if err := d.Rolie[i].Validate(); err != nil {
			return nil
		}
	}
	return nil
}

// Validate checks if the provider metadata is valid.
// Returns an error if the validation fails otherwise nil.
func (pmd *ProviderMetadata) Validate() error {

	switch {
	case pmd.CanonicalURL == nil:
		return errors.New("canonical_url is mandatory")
	case pmd.LastUpdated == nil:
		return errors.New("last_updated is mandatory")
	case pmd.MetadataVersion == nil:
		return errors.New("metadata_version is mandatory")
	}

	if err := pmd.Publisher.Validate(); err != nil {
		return err
	}

	for i := range pmd.PGPKeys {
		if err := pmd.PGPKeys[i].Validate(); err != nil {
			return err
		}
	}

	for i := range pmd.Distributions {
		if err := pmd.Distributions[i].Validate(); err != nil {
			return err
		}
	}

	return nil
}

// SetLastUpdated updates the last updated timestamp of the feed.
func (pmd *ProviderMetadata) SetLastUpdated(t time.Time) {
	ts := TimeStamp(t.UTC())
	pmd.LastUpdated = &ts
}

// SetPGP sets the fingerprint and URL of the OpenPGP key
// of the feed. If the feed already has a key with
// given fingerprint the URL updated.
// If there is no such key it is append to the list of keys.
func (pmd *ProviderMetadata) SetPGP(fingerprint, url string) {
	for i := range pmd.PGPKeys {
		if pmd.PGPKeys[i].Fingerprint == Fingerprint(fingerprint) {
			pmd.PGPKeys[i].URL = &url
			return
		}
	}
	pmd.PGPKeys = append(pmd.PGPKeys, PGPKey{
		Fingerprint: Fingerprint(fingerprint),
		URL:         &url,
	})
}

// NewProviderMetadata creates a new provider with the given URL.
// Valid default values are set and the feed is considered to
// be updated recently.
func NewProviderMetadata(canonicalURL string) *ProviderMetadata {
	pm := new(ProviderMetadata)
	pm.Defaults()
	pm.SetLastUpdated(time.Now())
	cu := ProviderURL(canonicalURL)
	pm.CanonicalURL = &cu
	return pm
}

// Save saves a metadata provider to a writer.
func (pmd *ProviderMetadata) Save(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(pmd)
}

// LoadProviderMetadata loads a metadata provider from a reader.
func LoadProviderMetadata(r io.Reader) (*ProviderMetadata, error) {

	var pmd ProviderMetadata
	dec := json.NewDecoder(r)
	if err := dec.Decode(&pmd); err != nil {
		return nil, err
	}

	if err := pmd.Validate(); err != nil {
		return nil, err
	}

	// Set defaults.
	pmd.Defaults()

	return &pmd, nil
}