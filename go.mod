module github.com/vtpl1/avsdk

go 1.22

retract (
    [v0.0.0, v0.1.0] // Retract all from v3
)

require github.com/pion/sdp/v3 v3.0.10

require github.com/pion/randutil v0.1.0 // indirect
