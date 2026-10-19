module github.com/firede/mygo-platinum/diagram

go 1.27.1

require (
	github.com/egoist/mygo v0.5.5
	github.com/firede/limu v0.0.0
	github.com/firede/limu/mygoui v0.0.0
	github.com/firede/mygo-platinum v0.0.0
)

require (
	github.com/ebitengine/purego v0.11.1 // indirect
	github.com/firede/platinum-fonts v0.2.0 // indirect
	github.com/go-text/typesetting v0.3.5 // indirect
	golang.org/x/image v0.46.0 // indirect
	golang.org/x/text v0.43.0 // indirect
)

// limu is not published yet: it is read from a checkout beside this one.
replace (
	github.com/firede/limu => ../../limu
	github.com/firede/limu/mygoui => ../../limu/mygoui
	github.com/firede/mygo-platinum => ../
)
