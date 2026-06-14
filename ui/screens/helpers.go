package screens

func overlayDialog(base, overlay string, width, height int) string {
	if overlay == "" {
		return base
	}
	return overlay
}
