// Task #2, #3, #4: Express Edge Server with MongoDB and Analytics
import express from 'express';
import type { Request, Response } from 'express';
import mongoose from 'mongoose';
import cors from 'cors';
import { ArduinoSerial } from './arduinoSerial.js';
import SensorData from './models/SensorData.js';
import Rule from './models/Rule.js';

// Test Bun global access
console.log(`🚀 Running on Bun ${Bun.version}`);

const app = express();
const PORT = process.env.PORT || 3000;
const MONGODB_URI = process.env.MONGODB_URI || 'mongodb://localhost:27017/smarthome';

// Middleware
app.use(cors());
app.use(express.json());

// Arduino serial connection
let arduino: ArduinoSerial | null = null;

// Connect to MongoDB
mongoose.connect(MONGODB_URI)
  .then(() => {
    console.log('✓ Connected to MongoDB');
    initializeDefaultRules();
  })
  .catch((err) => {
    console.error('MongoDB connection error:', err);
  });

// Initialize default rules
async function initializeDefaultRules() {
  try {
    const count = await Rule.countDocuments();
    if (count === 0) {
      const defaultRules = [
        {
          name: 'Gas Danger Alert',
          sensor: 'gas',
          operator: '>',
          threshold: 700,
          action: 'buzzer_on',
          enabled: true,
          description: 'Trigger buzzer when gas level exceeds danger threshold'
        },
        {
          name: 'Rain Detection - Close Window',
          sensor: 'water',
          operator: '>',
          threshold: 800,
          action: 'window_close',
          enabled: true,
          description: 'Automatically close window when rain is detected'
        },
        {
          name: 'Low Soil Moisture Alert',
          sensor: 'soil',
          operator: '>',
          threshold: 50,
          action: 'buzzer_on',
          enabled: true,
          description: 'Alert when soil moisture is too low'
        },
        {
          name: 'Auto Light - Low Light Detection',
          sensor: 'light',
          operator: '<',
          threshold: 300,
          action: 'white_light_on',
          enabled: true,
          description: 'Automatically turn on LED when light level is low'
        }
      ];

      await Rule.insertMany(defaultRules);
      console.log('✓ Default rules initialized');
    }
  } catch (error) {
    console.error('Error initializing rules:', error);
  }
}

// API Routes

// Health check
app.get('/api/health', (req: Request, res: Response) => {
  res.json({ 
    status: 'ok', 
    arduinoConnected: arduino !== null,
    mongodbConnected: mongoose.connection.readyState === 1
  });
});

// List available serial ports
app.get('/api/serial/ports', async (req: Request, res: Response) => {
  try {
    const ports = await ArduinoSerial.listPorts();
    res.json(ports);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Connect to Arduino
app.post('/api/serial/connect', async (req: Request, res: Response) => {
  try {
    const { port, baudRate } = req.body;
    
    if (!port) {
      return res.status(400).json({ error: 'Port is required' });
    }

    arduino = new ArduinoSerial();
    await arduino.connect(port, 9600);
    
    res.json({ message: 'Connected to Arduino', port });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Disconnect from Arduino
app.post('/api/serial/disconnect', async (req: Request, res: Response) => {
  try {
    if (arduino) {
      await arduino.disconnect();
      arduino = null;
    }
    res.json({ message: 'Disconnected from Arduino' });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Send command to Arduino
app.post('/api/serial/command', (req: Request, res: Response) => {
  try {
    const { command } = req.body;
    
    if (!arduino) {
      return res.status(400).json({ error: 'Arduino not connected' });
    }

    if (!command) {
      return res.status(400).json({ error: 'Command is required' });
    }

    arduino.sendCommand(command);
    res.json({ message: 'Command sent', command });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Get current sensor data
app.get('/api/sensors/current', (req: Request, res: Response) => {
  try {
    if (!arduino) {
      return res.status(400).json({ error: 'Arduino not connected' });
    }

    const data = arduino.getCurrentData();
    res.json(data);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Get current actuator states
app.get('/api/actuators/states', (req: Request, res: Response) => {
  try {
    if (!arduino) {
      return res.status(400).json({ error: 'Arduino not connected' });
    }

    const states = arduino.getActuatorStates();
    res.json(states);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Manually set actuator state (for synchronization)
app.post('/api/actuators/sync', (req: Request, res: Response) => {
  try {
    if (!arduino) {
      return res.status(400).json({ error: 'Arduino not connected' });
    }

    const { actuator, value } = req.body;
    
    if (!actuator || value === undefined) {
      return res.status(400).json({ error: 'Actuator and value are required' });
    }

    arduino.setActuatorState(actuator, value);
    res.json({ message: 'Actuator state updated', actuator, value });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Get historical sensor data with pagination
app.get('/api/sensors/history', async (req: Request, res: Response) => {
  try {
    const { limit = 100, skip = 0, startDate, endDate } = req.query;
    
    let query: any = {};
    
    if (startDate || endDate) {
      query.timestamp = {};
      if (startDate) query.timestamp.$gte = new Date(startDate as string);
      if (endDate) query.timestamp.$lte = new Date(endDate as string);
    }

    const data = await SensorData
      .find(query)
      .sort({ timestamp: -1 })
      .limit(Number(limit))
      .skip(Number(skip));

    const total = await SensorData.countDocuments(query);

    res.json({ data, total, limit: Number(limit), skip: Number(skip) });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Task #5: Analytics endpoints
app.get('/api/analytics/statistics', async (req: Request, res: Response) => {
  try {
    const { sensor, hours = 24 } = req.query;
    
    const startTime = new Date();
    startTime.setHours(startTime.getHours() - Number(hours));

    const pipeline: any[] = [
      { $match: { timestamp: { $gte: startTime } } }
    ];

    if (sensor && typeof sensor === 'string') {
      const validSensors = ['light', 'gas', 'soil', 'water'];
      if (!validSensors.includes(sensor)) {
        return res.status(400).json({ error: 'Invalid sensor type' });
      }

      pipeline.push({
        $group: {
          _id: null,
          mean: { $avg: `$${sensor}` },
          min: { $min: `$${sensor}` },
          max: { $max: `$${sensor}` },
          count: { $sum: 1 }
        }
      });
    } else {
      // Statistics for all sensors
      pipeline.push({
        $group: {
          _id: null,
          light_mean: { $avg: '$light' },
          light_min: { $min: '$light' },
          light_max: { $max: '$light' },
          gas_mean: { $avg: '$gas' },
          gas_min: { $min: '$gas' },
          gas_max: { $max: '$gas' },
          soil_mean: { $avg: '$soil' },
          soil_min: { $min: '$soil' },
          soil_max: { $max: '$soil' },
          water_mean: { $avg: '$water' },
          water_min: { $min: '$water' },
          water_max: { $max: '$water' },
          count: { $sum: 1 }
        }
      });
    }

    const result = await SensorData.aggregate(pipeline);
    
    res.json(result[0] || {});
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Get sensor trends (hourly averages)
app.get('/api/analytics/trends', async (req: Request, res: Response) => {
  try {
    const { hours = 24 } = req.query;
    
    const startTime = new Date();
    startTime.setHours(startTime.getHours() - Number(hours));

    const trends = await SensorData.aggregate([
      { $match: { timestamp: { $gte: startTime } } },
      {
        $group: {
          _id: {
            $dateToString: { format: '%Y-%m-%d %H:00', date: '$timestamp' }
          },
          light: { $avg: '$light' },
          gas: { $avg: '$gas' },
          soil: { $avg: '$soil' },
          water: { $avg: '$water' },
          count: { $sum: 1 }
        }
      },
      { $sort: { _id: 1 } }
    ]);

    res.json(trends);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Rules Management (Task #4)

// Get all rules
app.get('/api/rules', async (req: Request, res: Response) => {
  try {
    const rules = await Rule.find().sort({ createdAt: -1 });
    res.json(rules);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Create new rule
app.post('/api/rules', async (req: Request, res: Response) => {
  try {
    const rule = new Rule(req.body);
    await rule.save();
    res.status(201).json(rule);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Update rule
app.put('/api/rules/:id', async (req: Request, res: Response) => {
  try {
    const rule = await Rule.findByIdAndUpdate(
      req.params.id,
      { ...req.body, updatedAt: new Date() },
      { new: true }
    );
    
    if (!rule) {
      return res.status(404).json({ error: 'Rule not found' });
    }
    
    res.json(rule);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Delete rule
app.delete('/api/rules/:id', async (req: Request, res: Response) => {
  try {
    const rule = await Rule.findByIdAndDelete(req.params.id);
    
    if (!rule) {
      return res.status(404).json({ error: 'Rule not found' });
    }
    
    res.json({ message: 'Rule deleted' });
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Get recent alerts
app.get('/api/alerts', async (req: Request, res: Response) => {
  try {
    const { limit = 50 } = req.query;
    
    const alerts = await SensorData
      .find({ alerts: { $exists: true, $not: { $size: 0 } } })
      .sort({ timestamp: -1 })
      .limit(Number(limit))
      .select('alerts timestamp');

    res.json(alerts);
  } catch (error: any) {
    res.status(500).json({ error: error.message });
  }
});

// Start server
app.listen(PORT, () => {
  console.log(`🚀 Edge server running on http://localhost:${PORT}`);
});

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('\n🛑 Shutting down gracefully...');
  
  if (arduino) {
    await arduino.disconnect();
  }
  
  await mongoose.connection.close();
  process.exit(0);
});
