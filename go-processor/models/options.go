package models

type ProcessOptions struct {
	Format string // "jpg" or "png"
	Size int // Target size in pixels (square)
}

func (o *ProcessOptions) SetDefaults() {
	if o.Format == "" {
		o.Format = "jpg"
	}
	if o.Size == 0 {
		o.Size = 512
	}
}
