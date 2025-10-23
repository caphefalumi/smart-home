// Task #3: MongoDB Schema for storing sensor data
import mongoose, { Schema, Document } from 'mongoose';

export interface ISensorData extends Document {
  light: number;
  gas: number;
  soil: number;
  water: number;
  infrared: number;
  timestamp: Date;
  alerts: string[];
}

const SensorDataSchema: Schema = new Schema({
  light: { type: Number, required: true },
  gas: { type: Number, required: true },
  soil: { type: Number, required: true },
  water: { type: Number, required: true },
  infrared: { type: Number, required: false, default: 0 },
  timestamp: { type: Date, default: Date.now },
  alerts: [{ type: String }]
});

// Create indexes for efficient querying
SensorDataSchema.index({ timestamp: -1 });

export default mongoose.model<ISensorData>('SensorData', SensorDataSchema);
