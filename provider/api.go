package provider

const COS = "cos" // cos 名称
const OSS = "oss" // oss 名称

// A Provider describes an interface for providing files
type Provider interface {
	PutFile(key string, filepath string) error
	GetFile(key string, filepath string) error
	List(prefix string, marker string) []string
}
