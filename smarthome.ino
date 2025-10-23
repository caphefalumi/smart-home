#include <Servo.h>

// Servo objects - matching original names
Servo servo_10;  // Window servo (digital 10)
Servo servo_9;   // Door servo (digital 9)

// Sensor variables
volatile int gas;
volatile int light;
volatile int soil;
volatile int water;
volatile int infrar;

// Button variables
volatile int button1;  // Password button (digital 4)
volatile int button2;  // Confirm button (digital 8)

// Control variables
String command;

void setup() {
  Serial.begin(9600);
  
  // Initialize servos - matching original setup
  servo_9.attach(9);
  servo_10.attach(10);
  servo_9.write(0);
  servo_10.write(0);
  delay(300);
  
  // Initialize pins - EXACTLY matching original code
  pinMode(7, OUTPUT);    // Fan direction
  pinMode(6, OUTPUT);    // Fan PWM
  digitalWrite(7, HIGH);
  digitalWrite(6, HIGH);
  
  pinMode(4, INPUT);     // Password button
  pinMode(8, INPUT);     // Confirm button
  pinMode(2, INPUT);     // Infrared sensor
  pinMode(3, OUTPUT);    // Buzzer
  pinMode(A0, INPUT);    // Gas sensor
  pinMode(A1, INPUT);    // Light sensor
  pinMode(13, OUTPUT);   // White light (LED)
  pinMode(A3, INPUT);    // Water sensor
  pinMode(A2, INPUT);    // Soil sensor
  pinMode(12, OUTPUT);   // Relay
  pinMode(5, OUTPUT);    // Yellow light (LED2)
}

void loop() {
  // Read all sensor values
  readSensors();
  // Read button states
  readButtons();
  // Send sensor data to serial
  sendSensorData();
  // Check for incoming serial commands
  while (Serial.available() > 0) {
    command = Serial.readStringUntil('\n');
    command.trim();
    if (command.length() > 0) {
      processActuatorCommands();
      Serial.println("ACK"); // <— confirmation back to Node.js
    }
  }

  delay(100);
}

void readSensors() {
  gas = analogRead(A0);
  light = analogRead(A1);
  soil = analogRead(A2);
  water = analogRead(A3);
  infrar = digitalRead(2);
}

void readButtons() {
  button1 = digitalRead(4);
  button2 = digitalRead(8);
}

void sendSensorData() {
  Serial.print("GAS:");
  Serial.print(gas);
  Serial.print(",LIGHT:");
  Serial.print(light);
  Serial.print(",SOIL:");
  Serial.print(soil);
  Serial.print(",WATER:");
  Serial.print(water);
  Serial.print(",INFRAR:");
  Serial.print(infrar);
  Serial.print(",BTN1:");
  Serial.print(button1);
  Serial.print(",BTN2:");
  Serial.println(button2);
}

void processActuatorCommands() {
  // White light control (Digital 13 - original case 'a'/'b')
  if (command.startsWith("white_light_on")) {
    digitalWrite(13, HIGH);
    Serial.println("White light ON");
  }
  else if (command.startsWith("white_light_off")) {
    digitalWrite(13, LOW);
    Serial.println("White light OFF");
  }
  
  // Yellow light control (Digital 5 - original case 'p'/'q')
  else if (command.startsWith("yellow_light_on")) {
    digitalWrite(5, HIGH);
    Serial.println("Yellow light ON");
  }
  else if (command.startsWith("yellow_light_off")) {
    digitalWrite(5, LOW);
    Serial.println("Yellow light OFF");
  }
  
  // Relay control (Digital 12 - original case 'c'/'d')
  else if (command.startsWith("relay_on")) {
    digitalWrite(12, HIGH);
    Serial.println("Relay ON");
  }
  else if (command.startsWith("relay_off")) {
    digitalWrite(12, LOW);
    Serial.println("Relay OFF");
  }
  
  // Door servo control (Digital 9 - original case 'l'/'m')
  else if (command.startsWith("door_open")) {
    servo_9.write(180);
    Serial.println("Door opened");
    delay(500);
  }
  else if (command.startsWith("door_close")) {
    servo_9.write(0);
    Serial.println("Door closed");
    delay(500);
  }
  else if (command.startsWith("door_angle=")) {
    String angleStr = command.substring(11);
    int angle = angleStr.toInt();
    servo_9.write(angle);
    Serial.print("Door angle: ");
    Serial.println(angle);
    delay(300);
  }
  
  // Window servo control (Digital 10 - original case 'n'/'o')
  else if (command.startsWith("window_open")) {
    servo_10.write(180);
    Serial.println("Window opened");
    delay(500);
  }
  else if (command.startsWith("window_close")) {
    servo_10.write(0);
    Serial.println("Window closed");
    delay(500);
  }
  else if (command.startsWith("window_angle=")) {
    String angleStr = command.substring(13);
    int angle = angleStr.toInt();
    servo_10.write(angle);
    Serial.print("Window angle: ");
    Serial.println(angle);
    delay(300);
  }
  
  // Fan control (Digital 7 & 6 - original case 'r'/'s')
  else if (command.startsWith("fan_on")) {
    digitalWrite(7, LOW);
    digitalWrite(6, HIGH);
    Serial.println("Fan ON");
  }
  else if (command.startsWith("fan_off")) {
    digitalWrite(7, LOW);
    digitalWrite(6, LOW);
    Serial.println("Fan OFF");
  }
  else if (command.startsWith("fan_speed=")) {
    String speedStr = command.substring(10);
    int speed = speedStr.toInt();
    digitalWrite(7, LOW);
    analogWrite(6, speed);
    Serial.print("Fan speed: ");
    Serial.println(speed);
  }
  
  // Yellow light PWM control (Digital 5 - original case 'v')
  else if (command.startsWith("yellow_light_pwm=")) {
    String pwmStr = command.substring(17);
    int pwm = pwmStr.toInt();
    analogWrite(5, pwm);
    Serial.print("Yellow light PWM: ");
    Serial.println(pwm);
  }
  
  // Buzzer control (Digital 3)
  else if (command.startsWith("buzzer_on")) {
    tone(3, 1000);
    Serial.println("Buzzer ON");
  }
  else if (command.startsWith("buzzer_off")) {
    noTone(3);
    Serial.println("Buzzer OFF");
  }
  else {
    Serial.print("Unknown command: ");
    Serial.println(command);
  }
}