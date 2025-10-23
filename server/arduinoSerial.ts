// Task #2: Serial Communication Handler for Arduino
import { SerialPort } from 'serialport';
import { ReadlineParser } from '@serialport/parser-readline';
import SensorData from './models/SensorData.js';
import Rule from './models/Rule.js';

export interface SensorReading {
  gas: number;
  light: number;
  soil: number;
  water: number;
  infrar: number;
  btn1: number;
  btn2: number;
  timestamp: Date;
}

export interface ActuatorStates {
  white_light: boolean;
  yellow_light: boolean;
  relay: boolean;
  door_angle: number;
  window_angle: number;
  fan: boolean;
  fan_speed: number;
  buzzer: boolean;
}

export class ArduinoSerial {
  private port: SerialPort | null = null;
  private parser: ReadlineParser | null = null;
  private currentData: SensorReading | null = null;
  private actuatorStates: ActuatorStates = {
    white_light: false,
    yellow_light: false,
    relay: false,
    door_angle: 0,
    window_angle: 0,
    fan: false,
    fan_speed: 0,
    buzzer: false
  };
  private dataBuffer: SensorReading[] = [];
  private saveInterval: NodeJS.Timeout | null = null;
  private isReady = true;
  private commandQueue: string[] = [];
  constructor() {
    // Start saving data every 2 seconds
    this.saveInterval = setInterval(() => {
      this.saveBufferedData();
    }, 2000);
  }

  static async listPorts(): Promise<any[]> {
    const { SerialPort } = await import('serialport');
    return SerialPort.list();
  }

  async connect(portPath: string, baudRate: number = 9600): Promise<void> {
    try {
      if (this.port && this.port.isOpen) {
        console.warn('Port already open');
        return;
      }

      // Open serial port
      this.port = new SerialPort({ path: portPath, baudRate });

      // Wait 2 seconds for Arduino to reset after port open
      await new Promise(res => setTimeout(res, 2000));

      // Set up parser for newline-delimited data
      this.parser = this.port.pipe(new ReadlineParser({ delimiter: '\n' }));

      // Set up listeners
      this.setupListeners();

      console.log(`✓ Connected to Arduino on ${portPath}`);
    } catch (error) {
      console.error('Failed to connect to Arduino:', error);
      throw error;
    }
  }

  async disconnect(): Promise<void> {
    if (this.saveInterval) {
      clearInterval(this.saveInterval);
      this.saveInterval = null;
    }
    
    if (this.port && this.port.isOpen) {
      await new Promise<void>((resolve) => {
        this.port!.close(() => {
          console.log('✓ Disconnected from Arduino');
          resolve();
        });
      });
    }
    
    this.port = null;
    this.parser = null;
  }

  private setupListeners(): void {
    if (!this.parser || !this.port) return;

    // Parser listens for each line from Arduino
    this.parser.on('data', (line: string) => {
      const trimmed = line.trim();
      console.log('←', trimmed);

      // ACK handling for command queue
      if (trimmed === 'ACK') {
        this.isReady = true;

        // Send next command if queued
        if (this.commandQueue.length > 0) {
          const nextCommand = this.commandQueue.shift()!;
          setTimeout(() => this.sendCommand(nextCommand), 100); // slight delay
        }
        return;
      }

      // Process sensor or actuator data
      this.handleIncomingData(trimmed);
    });

    // Handle parser errors
    this.parser.on('error', (error) => {
      console.error('Serial parser error:', error);
    });

    // Handle port-level errors
    this.port.on('error', (error) => {
      console.error('Serial port error:', error);
    });
  }


  private handleIncomingData(line: string): void {
    try {
      // Parse Arduino sensor data format: "GAS:123,LIGHT:456,SOIL:789,WATER:101,INFRAR:1,BTN1:0,BTN2:1"
      if (line.includes('GAS:')) {
        const sensorData = this.parseSensorData(line);
        if (sensorData) {
          this.currentData = sensorData;
          this.dataBuffer.push(sensorData);
          
          // Keep buffer size manageable
          if (this.dataBuffer.length > 100) {
            this.dataBuffer = this.dataBuffer.slice(-50);
          }
          
          // Evaluate rules for edge analytics
          this.evaluateRules(sensorData);
        }
      }
      // Handle actuator acknowledgments
      else if (line.includes('light ON') || line.includes('light OFF') || 
               line.includes('opened') || line.includes('closed') ||
               line.includes('Fan ON') || line.includes('Fan OFF') ||
               line.includes('Relay ON') || line.includes('Relay OFF') ||
               line.includes('Buzzer ON') || line.includes('Buzzer OFF')) {
        this.updateActuatorStateFromResponse(line);
      }
    } catch (error) {
      console.error('Error handling incoming data:', error);
    }
  }

  private parseSensorData(line: string): SensorReading | null {
    try {
      const parts = line.split(',');
      const data: any = {};
      
      parts.forEach(part => {
        const [key, value] = part.split(':');
        if (key && value !== undefined) {
          data[key.toLowerCase()] = parseInt(value, 10);
        }
      });

      if (data.gas !== undefined && data.light !== undefined && 
          data.soil !== undefined && data.water !== undefined) {
        return {
          gas: data.gas,
          light: data.light,
          soil: data.soil,
          water: data.water,
          infrar: data.infrar || 0,
          btn1: data.btn1 || 0,
          btn2: data.btn2 || 0,
          timestamp: new Date()
        };
      }
      
      return null;
    } catch (error) {
      console.error('Error parsing sensor data:', error);
      return null;
    }
  }

  private updateActuatorStateFromResponse(response: string): void {
    if (response.includes('White light ON')) {
      this.actuatorStates.white_light = true;
    } else if (response.includes('White light OFF')) {
      this.actuatorStates.white_light = false;
    } else if (response.includes('Yellow light ON')) {
      this.actuatorStates.yellow_light = true;
    } else if (response.includes('Yellow light OFF')) {
      this.actuatorStates.yellow_light = false;
    } else if (response.includes('Relay ON')) {
      this.actuatorStates.relay = true;
    } else if (response.includes('Relay OFF')) {
      this.actuatorStates.relay = false;
    } else if (response.includes('Door opened')) {
      this.actuatorStates.door_angle = 180;
    } else if (response.includes('Door closed')) {
      this.actuatorStates.door_angle = 0;
    } else if (response.includes('Window opened')) {
      this.actuatorStates.window_angle = 180;
    } else if (response.includes('Window closed')) {
      this.actuatorStates.window_angle = 0;
    } else if (response.includes('Fan ON')) {
      this.actuatorStates.fan = true;
    } else if (response.includes('Fan OFF')) {
      this.actuatorStates.fan = false;
    } else if (response.includes('Buzzer ON')) {
      this.actuatorStates.buzzer = true;
    } else if (response.includes('Buzzer OFF')) {
      this.actuatorStates.buzzer = false;
    } else if (response.includes('Door angle:')) {
      const angle = parseInt(response.split(':')[1]?.trim() || '0', 10);
      this.actuatorStates.door_angle = angle;
    } else if (response.includes('Window angle:')) {
      const angle = parseInt(response.split(':')[1]?.trim() || '0', 10);
      this.actuatorStates.window_angle = angle;
    } else if (response.includes('Fan speed:')) {
      const speed = parseInt(response.split(':')[1]?.trim() || '0', 10);
      this.actuatorStates.fan_speed = speed;
    }
  }

  private async evaluateRules(sensorData: SensorReading): Promise<void> {
    try {
      const rules = await Rule.find({ enabled: true });
      
      for (const rule of rules) {
        const sensorValue = sensorData[rule.sensor as keyof SensorReading] as number;
        let triggered = false;
        
        switch (rule.operator as any) {
          case '>':
            triggered = sensorValue > rule.threshold;
            break;
          case '<':
            triggered = sensorValue < rule.threshold;
            break;
          case '>=':
            triggered = sensorValue >= rule.threshold;
            break;
          case '<=':
            triggered = sensorValue <= rule.threshold;
            break;
          case '==':
            triggered = sensorValue === rule.threshold;
            break;
        }
        
        if (triggered) {
          console.log(`🔥 Rule triggered: ${rule.name} (${rule.sensor}: ${sensorValue} ${rule.operator} ${rule.threshold})`);
          await this.executeAction(rule.action, sensorData);
        }
      }
    } catch (error) {
      console.error('Error evaluating rules:', error);
    }
  }

  private async executeAction(action: string, sensorData: SensorReading): Promise<void> {
    try {
      switch (action) {
        case 'white_light_on':
          if (!this.actuatorStates.white_light) {
            this.sendCommand('white_light_on');
          }
          break;
        case 'white_light_off':
          if (this.actuatorStates.white_light) {
            this.sendCommand('white_light_off');
          }
          break;
        case 'yellow_light_on':
          if (!this.actuatorStates.yellow_light) {
            this.sendCommand('yellow_light_on');
          }
          break;
        case 'yellow_light_off':
          if (this.actuatorStates.yellow_light) {
            this.sendCommand('yellow_light_off');
          }
          break;
        case 'relay_on':
          if (!this.actuatorStates.relay) {
            this.sendCommand('relay_on');
          }
          break;
        case 'relay_off':
          if (this.actuatorStates.relay) {
            this.sendCommand('relay_off');
          }
          break;
        case 'door_open':
          if (this.actuatorStates.door_angle !== 180) {
            this.sendCommand('door_open');
          }
          break;
        case 'door_close':
          if (this.actuatorStates.door_angle !== 0) {
            this.sendCommand('door_close');
          }
          break;
        case 'window_open':
          if (this.actuatorStates.window_angle !== 180) {
            this.sendCommand('window_open');
          }
          break;
        case 'window_close':
          if (this.actuatorStates.window_angle !== 0) {
            this.sendCommand('window_close');
          }
          break;
        case 'fan_on':
          if (!this.actuatorStates.fan) {
            this.sendCommand('fan_on');
          }
          break;
        case 'fan_off':
          if (this.actuatorStates.fan) {
            this.sendCommand('fan_off');
          }
          break;
        case 'buzzer_on':
          this.sendCommand('buzzer_on');
          // Auto turn off buzzer after 2 seconds
          setTimeout(() => {
            this.sendCommand('buzzer_off');
          }, 2000);
          break;
        case 'buzzer_off':
          if (this.actuatorStates.buzzer) {
            this.sendCommand('buzzer_off');
          }
          break;
        default:
          console.warn(`Unknown action: ${action}`);
      }
      
      // Save alert to database
      if (this.currentData) {
        const alertData = {
          ...this.currentData,
          alerts: [`Action executed: ${action}`]
        };
        
        const sensorDoc = new SensorData(alertData);
        await sensorDoc.save();
      }
    } catch (error) {
      console.error('Error executing action:', error);
    }
  }

  sendCommand(command: string): void {
    if (!this.port || !this.port.isOpen) {
      console.warn('Cannot send command: Arduino not connected');
      return;
    }

    // Queue command if Arduino is busy
    if (!this.isReady) {
      this.commandQueue.push(command);
      console.log(`⏳ Queued: ${command}`);
      return;
    }

    // Send immediately if ready
    this.isReady = false;
    this.port.write(command + '\n', (err) => {
      if (err) console.error('Error writing to port:', err);
      else console.log(`→ Sent: ${command}`);
    });
  }


  getCurrentData(): SensorReading | null {
    return this.currentData;
  }

  getActuatorStates(): ActuatorStates {
    return { ...this.actuatorStates };
  }

  setActuatorState(actuator: string, value: any): void {
    if (actuator in this.actuatorStates) {
      (this.actuatorStates as any)[actuator] = value;
    }
  }

  private async saveBufferedData(): Promise<void> {
    if (this.dataBuffer.length === 0) return;
    
    try {
      const dataToSave = this.dataBuffer.splice(0);
      
      if (dataToSave.length > 0) {
        await SensorData.insertMany(dataToSave);
        console.log(`💾 Saved ${dataToSave.length} sensor readings to database`);
      }
    } catch (error) {
      console.error('Error saving sensor data:', error);
    }
  }
}
