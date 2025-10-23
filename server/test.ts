import { SerialPort } from 'serialport';

const { ReadlineParser } = require('@serialport/parser-readline')
const port = new SerialPort({ path: 'COM7', baudRate: 9600 });
const parser = port.pipe(new ReadlineParser())
port.on('open', () => console.log('Port opened'));
port.on('error', (err) => console.error('Serial Port Error:', err));
port.on('data', (data) => {
  console.log('Raw Data:', data);
});
parser.on('data', (data) => {
  console.log('Parsed Data:', data);
});
