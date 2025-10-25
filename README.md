# Introduction

A modern smart home automation system with real-time sensor monitoring, rule-based automation, and device control.

## Demo


[![Demo Video](https://i.ibb.co/1YGZjtnV/Demo.png)](https://youtu.be/VcMubtFll-w)

- View live sensor data (gas, light, soil, water, infrared)
- Set up automation rules (e.g., "Turn on fan if gas > 100")
- Control actuators (lights, fan, relay, buzzer, etc.)
- See analytics and trends

## Purpose

This project helps homeowners and researchers monitor and automate their environment using affordable sensors and actuators. It brings value by:

- Improving safety (gas leak alerts, water detection)
- Saving energy (automated lighting/fan control)
- Enabling remote monitoring and control

## Installation

1. Clone the repository:

   ```sh
   git clone https://github.com/caphefalumi/smart-home.git
   ```

2. Install dependencies for the frontend:

   ```sh
   cd smart-home
   bun install
   ```

3. Install Go dependencies for the backend:

   ```sh
   cd server
   go mod tidy
   ```

4. Configure your MongoDB URI in `server/config/config.go`.

5. Start the backend:

   ```sh
   go run main.go
   ```

6. Start the frontend:

   ```sh
   bun dev
   ```

## Quick Start

- Open [http://localhost:5173](http://localhost:5173) in your browser
- Connect your Arduino via serial port
- Add rules and watch automation in action

## Libraries Used

- [Vue.js](https://vuejs.org/) (frontend)
- [Vuetify](https://vuetifyjs.com/) (UI components)
- [Vite](https://vitejs.dev/) (build tool)
- [Go](https://golang.org/) (backend)
- [Gin](https://gin-gonic.com/) (web framework)
- [MongoDB](https://www.mongodb.com/) (database)

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
