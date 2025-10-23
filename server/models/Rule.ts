// Task #4: Conditional Rules Schema
import mongoose, { Schema, Document } from 'mongoose';

export interface IRule extends Document {
  name: string;
  sensor: 'gas' | 'light' | 'soil' | 'water';
  operator: '>' | '<' | '>=' | '<=' | '==';
  threshold: number;
  action: string;
  enabled: boolean;
  description?: string;
  createdAt: Date;
  updatedAt: Date;
}

const RuleSchema: Schema = new Schema({
  name: { type: String, required: true },
  sensor: { 
    type: String, 
    required: true,
    enum: ['gas', 'light', 'soil', 'water']
  },  operator: { 
    type: String, 
    required: true,
    enum: ['>', '<', '>=', '<=', '==']
  },
  threshold: { type: Number, required: true },
  action: { type: String, required: true },
  enabled: { type: Boolean, default: true },
  description: { type: String },
  createdAt: { type: Date, default: Date.now },
  updatedAt: { type: Date, default: Date.now }
});

export default mongoose.model<IRule>('Rule', RuleSchema);
