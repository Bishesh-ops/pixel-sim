# Falling Sand Engine

A high-performance, 2D granular physics and thermodynamics simulation built in **Go** using the **Ebitengine** library. This project features a custom cellular automata engine capable of simulating fluid dynamics, heat diffusion, and material state changes.

## 🚀 Features

* **15+ Unique Materials:** Includes solids (Sand, Stone, Wood), liquids (Water, Acid, Lava), gases (Smoke, Steam, Gas), and energy (VoltBolt).
* **Thermodynamics Engine:** A secondary heat-map grid tracks temperature diffusion across the environment.
* **Dynamic State Changes:** Materials react to heat (e.g., Sand melts into Glass at high temperatures, Water boils into Steam).
* **Biological Growth:** Plants consume Water to grow and spread dynamically through the grid.
* **Combustion & Explosions:** Realistic fire spread through flammable materials and radius-based TNT force.
* **Custom UI Framework:** Integrated sidebar for material selection, brush size control, and real-time tooltips.

## 🛠️ Technical Stack

* **Language:** Go (Golang)
* **Graphics Library:** [Ebitengine](https://ebitengine.org/)
* **Concepts:** Cellular Automata, Heat Diffusion, Probability-based State Machines, and Pixel-Buffer Manipulation.

## 🎮 Controls

| Input | Action |
| :--- | :--- |
| **Left Click** | Spawn Selected Material |
| **Right Click** | Erase (Empty) |
| **Mouse Wheel** | Adjust Brush Size |
| **Spacebar** | Pause/Resume Simulation |
| **'R' Key** | Reset Grid (Clear All) |
| **'S' Key** | Frame-by-frame Step (when paused) |
| **Escape** | Exit Application |

## 🏗️ Architecture

The engine utilizes a **Double-Buffered Grid** approach to minimize directional bias. Each frame, the simulation alternates the sweep direction (Left-to-Right vs. Right-to-Left) to ensure fluids and gases flow naturally across the $320 \times 240$ internal resolution.

### Thermodynamics Logic
Heat is calculated using a 4-point average diffusion algorithm:

$$T_{new} = \frac{T_{up} + T_{down} + T_{left} + T_{right}}{4} \times \text{dissipation}$$



## 🏁 Getting Started

1.  **Ensure you have Go installed** (Version 1.18+ recommended).
2.  **Clone the repository:**
    ```bash
    git clone [https://github.com/your-username/pixel-sim.git](https://github.com/your-username/pixel-sim.git)
    cd pixel-sim
    ```
3.  **Install dependencies:**
    ```bash
    go mod tidy
    ```
4.  **Run the simulation:**
    ```bash
    go run main.go
    ```

---

### About the Developer
**Bishesh Shrestha** B.S. in Computer Science with a Minor in AI from Dakota State University.  
Currently focused on Software Engineering and Data Engineering roles.
