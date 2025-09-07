package feature

// Flag is named such that checking for a feature uses `feature.Flag.Enabled(feature.ExampleFeature)`.
var Flag = New()

// flag names are written in kebab-case
const (
	BackendErrorRedesign    FlagName = "backend-error-redesign"
	DeviceIDForHardlinks    FlagName = "device-id-for-hardlinks"
	ExplicitS3AnonymousAuth FlagName = "explicit-s3-anonymous-auth"
	SafeForgetKeepTags      FlagName = "safe-forget-keep-tags"
	ScriptModeForLs         FlagName = "script-mode-for-ls"
)

func init() {
	Flag.SetFlags(map[FlagName]FlagDesc{
		BackendErrorRedesign:    {Type: Beta, Description: "enforce timeouts for stuck HTTP requests and use new backend error handling design."},
		DeviceIDForHardlinks:    {Type: Alpha, Description: "store deviceID only for hardlinks to reduce metadata changes for example when using btrfs subvolumes. Will be removed in a future restic version after repository format 3 is available"},
		ExplicitS3AnonymousAuth: {Type: Beta, Description: "forbid anonymous S3 authentication unless `-o s3.unsafe-anonymous-auth=true` is set"},
		SafeForgetKeepTags:      {Type: Beta, Description: "prevent deleting all snapshots if the tag passed to `forget --keep-tags tagname` does not exist"},
		ScriptModeForLs:         {Type: Alpha, Description: "scriptable interface for querying repo contents through `restic ls --script-mode`"},
	})
}
