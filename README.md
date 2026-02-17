# Go Falling Sand

A granular physics sandbox written in Go using Ebitengine.I built this to dive into cellular automata, fluid dynamics, and pixel manipulation. It simulates 15+ elements interacting in real-time—liquids flow, gases rise, fire spreads, and plants grow (until you burn them down).

##  Features
### The Elements
Play with solids (Sand, Stone, Wood), liquids (Water, Acid, Lava), gases (Smoke, Steam), and even electricity (Volt).

### Reactions
Materials actually interact.Water boils into Steam when near Lava.Acid dissolves Stone and Wood.Sand melts into Glass if it gets hot enough.

### Thermodynamics
There's a background heat map calculating temperature diffusion. Things don't just change state randomly; they have to get hot first.

### Explosives
TNT creates pressure waves and destroys the environment.

### Custom UI
I wrote a custom sidebar to handle element selection and brush resizing without relying on heavy UI libraries.

##  How it Works
The engine runs on a fixed $320 \times 240$ grid, scaled up for the window.To keep the simulation stable and prevent "directional bias" (where sand piles up weirdly on one side), the engine alternates the scan direction every frame (Left-to-Right $\leftrightarrow$ Right-to-Left).

### The Heat Model
Temperature isn't just a static value; it diffuses using a 4-point average. This creates natural gradients rather than blocky heat zones:

$$T_{new} = \frac{T_{up} + T_{down} + T_{left} + T_{right}}{4} \times \text{decay}$$

#  ControlsKeyActionLeft ClickSpawn MaterialRight ClickEraserScroll WheelChange Brush SizeSpacePause / ResumeRReset GridEscQuit

#  Running LocallyYou'll need Go 1.18+ installed.Bash
# Clone the repo
git clone https://github.com/your-username/pixel-sim.git

cd pixel-sim

# Tidy up dependencies
go mod tidy

# Run it
go run main.go

***Built by Bishesh ShresthaCS & AI at Dakota State University.I'm currently looking for roles in Software or Data Engineering. Feel free to check out my other repos!***
