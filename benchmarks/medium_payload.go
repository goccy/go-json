package benchmark

import "github.com/francoispqt/gojay"

// Response from Clearbit API. Size: 2.4kb
var MediumFixture = []byte(`{
	"person": {
	  "id": "d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
	  "name": {
		"fullName": "Leonid Bugaev",
		"givenName": "Leonid",
		"familyName": "Bugaev"
	  },
	  "email": "leonsbox@gmail.com",
	  "gender": "male",
	  "location": "Saint Petersburg, Saint Petersburg, RU",
	  "geo": {
		"city": "Saint Petersburg",
		"state": "Saint Petersburg",
		"country": "Russia",
		"lat": 59.9342802,
		"lng": 30.3350986
	  },
	  "bio": "Senior engineer at Granify.com",
	  "site": "http://flickfaver.com",
	  "avatar": "https://d1ts43dypk8bqh.cloudfront.net/v1/avatars/d50887ca-a6ce-4e59-b89f-14f0b5d03b03",
	  "employment": {
		"name": "www.latera.ru",
		"title": "Software Engineer",
		"domain": "gmail.com"
	  },
	  "facebook": {
		"handle": "leonid.bugaev"
	  },
	  "github": {
		"handle": "buger",
		"id": 14009,
		"avatar": "https://avatars.githubusercontent.com/u/14009?v=3",
		"company": "Granify",
		"blog": "http://leonsbox.com",
		"followers": 95,
		"following": 10
	  },
	  "twitter": {
		"handle": "flickfaver",
		"id": 77004410,
		"bio": null,
		"followers": 2,
		"following": 1,
		"statuses": 5,
		"favorites": 0,
		"location": "",
		"site": "http://flickfaver.com",
		"avatar": null
	  },
	  "linkedin": {
		"handle": "in/leonidbugaev"
	  },
	  "googleplus": {
		"handle": null
	  },
	  "angellist": {
		"handle": "leonid-bugaev",
		"id": 61541,
		"bio": "Senior engineer at Granify.com",
		"blog": "http://buger.github.com",
		"site": "http://buger.github.com",
		"followers": 41,
		"avatar": "https://d1qb2nb5cznatu.cloudfront.net/users/61541-medium_jpg?1405474390"
	  },
	  "klout": {
		"handle": null,
		"score": null
	  },
	  "foursquare": {
		"handle": null
	  },
	  "aboutme": {
		"handle": "leonid.bugaev",
		"bio": null,
		"avatar": null
	  },
	  "gravatar": {
		"handle": "buger",
		"urls": [
		],
		"avatar": "http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510",
		"avatars": [
		  {
			"url": "http://1.gravatar.com/avatar/f7c8edd577d13b8930d5522f28123510",
			"type": "thumbnail"
		  }
		]
	  },
	  "fuzzy": false
	},
	"company": null
  }`)

type CBAvatar struct {
	Url string
}

func (m *CBAvatar) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "avatars":
		return dec.AddString(&m.Url)
	}
	return nil
}
func (m *CBAvatar) NKeys() int {
	return 1
}

func (m *CBAvatar) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("url", m.Url)
}

func (m *CBAvatar) IsNil() bool {
	return m == nil
}

type Avatars []*CBAvatar

func (t *Avatars) UnmarshalJSONArray(dec *gojay.Decoder) error {
	avatar := CBAvatar{}
	*t = append(*t, &avatar)
	return dec.AddObject(&avatar)
}

func (m *Avatars) MarshalJSONArray(enc *gojay.Encoder) {
	for _, e := range *m {
		enc.AddObject(e)
	}
}
func (m *Avatars) IsNil() bool {
	return m == nil
}

type CBGravatar struct {
	Avatars Avatars
}

func (m *CBGravatar) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "avatars":
		return dec.AddArray(&m.Avatars)
	}
	return nil
}
func (m *CBGravatar) NKeys() int {
	return 1
}

func (m *CBGravatar) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddArrayKey("avatars", &m.Avatars)
}

func (m *CBGravatar) IsNil() bool {
	return m == nil
}

type CBGithub struct {
	Followers int
}

func (m *CBGithub) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "followers":
		return dec.AddInt(&m.Followers)
	}
	return nil
}

func (m *CBGithub) NKeys() int {
	return 1
}

func (m *CBGithub) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddIntKey("followers", m.Followers)
}

func (m *CBGithub) IsNil() bool {
	return m == nil
}

type CBName struct {
	FullName string `json:"fullName"`
}

func (m *CBName) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "fullName":
		return dec.AddString(&m.FullName)
	}
	return nil
}

func (m *CBName) NKeys() int {
	return 1
}

func (m *CBName) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddStringKey("fullName", m.FullName)
}

func (m *CBName) IsNil() bool {
	return m == nil
}

type CBPerson struct {
	Name     *CBName   `json:"name"`
	Github   *CBGithub `json:"github"`
	Gravatar *CBGravatar
}

func (m *CBPerson) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "name":
		m.Name = &CBName{}
		return dec.AddObject(m.Name)
	case "github":
		m.Github = &CBGithub{}
		return dec.AddObject(m.Github)
	case "gravatar":
		m.Gravatar = &CBGravatar{}
		return dec.AddObject(m.Gravatar)
	}
	return nil
}

func (m *CBPerson) NKeys() int {
	return 3
}

func (m *CBPerson) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddObjectKey("name", m.Name)
	enc.AddObjectKey("github", m.Github)
	enc.AddObjectKey("gravatar", m.Gravatar)
}

func (m *CBPerson) IsNil() bool {
	return m == nil
}

type MediumPayload struct {
	Person  *CBPerson `json:"person"`
	Company string    `json:"company"`
}

//easyjson:json
type MediumPayloadEasyJson struct {
	Person  *CBPerson `json:"person"`
	Company string    `json:"company"`
}

func (m *MediumPayload) UnmarshalJSONObject(dec *gojay.Decoder, key string) error {
	switch key {
	case "person":
		m.Person = &CBPerson{}
		return dec.AddObject(m.Person)
	case "company":
		dec.AddString(&m.Company)
	}
	return nil
}

func (m *MediumPayload) NKeys() int {
	return 2
}

func (m *MediumPayload) MarshalJSONObject(enc *gojay.Encoder) {
	enc.AddObjectKey("person", m.Person)
	// enc.AddStringKey("company", m.Company)
}

func (m *MediumPayload) IsNil() bool {
	return m == nil
}

func NewMediumPayload() *MediumPayload {
	return &MediumPayload{
		Company: "test",
		Person: &CBPerson{
			Name: &CBName{
				FullName: "test",
			},
			Github: &CBGithub{
				Followers: 100,
			},
			Gravatar: &CBGravatar{
				Avatars: Avatars{
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
				},
			},
		},
	}
}

func NewMediumPayloadEasyJson() *MediumPayloadEasyJson {
	return &MediumPayloadEasyJson{
		Company: "test",
		Person: &CBPerson{
			Name: &CBName{
				FullName: "test",
			},
			Github: &CBGithub{
				Followers: 100,
			},
			Gravatar: &CBGravatar{
				Avatars: Avatars{
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
					&CBAvatar{
						Url: "http://test.com",
					},
				},
			},
		},
	}
}
