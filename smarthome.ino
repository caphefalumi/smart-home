#include <Servo.h>
#include <Wire.h>
#include <LiquidCrystal_I2C.h>

// Set the communication address of I2C to 0x27, display 16 characters every line, two lines in total
LiquidCrystal_I2C mylcd(0x27, 16, 2);

// Servo objects - matching original names
Servo servo_10;  // Window servo (digital 10)
Servo servo_9;   // Door servo (digital 9)

// Sensor variables - use uint16_t instead of int to save space
uint16_t gas, light, soil, water;
uint8_t infrar;

// Button variables - use uint8_t instead of int
uint8_t button1, button2;
uint8_t btn1_num = 0, btn2_num = 0;

// Password variables - use char arrays instead of String
char pass[17] = "";     // Display string for LCD (max 16 chars + null)
char passwd[17] = "";   // Actual password string
uint8_t flag = 0, flag2 = 0, flag3 = 0; // Sensor flags

// Music variables
uint8_t tonepin = 3; // Set the signal end of passive buzzer to digital 3

// Non-blocking timing variables
unsigned long lastMusicNoteTime = 0;
uint8_t currentNote = 0;
bool isPlayingMusic = false;
uint8_t currentTuneType = 0; // 1 for birthday, 2 for ode to joy

// Timing control
unsigned long lastSensorSend = 0;
const uint16_t SENSOR_INTERVAL = 1000; // Send sensor data every 1 second
#define D0 -1
#define D1 262
#define D2 293
#define D3 329
#define D4 349
#define D5 392
#define D6 440
#define D7 494
#define M1 523
#define M2 586
#define M3 658
#define M4 697
#define M5 783
#define M6 879
#define M7 987
#define H1 1045
#define H2 1171
#define H3 1316
#define H4 1393
#define H5 1563
#define H6 1755
#define H7 1971
// Define sound frequencies - use const to save RAM
const uint16_t PROGMEM tune[] = {
  M3, M3, M4, M5,
  M5, M4, M3, M2,
  M1, M1, M2, M3,
  M3, M2, M2,
  M3, M3, M4, M5,
  M5, M4, M3, M2,
  M1, M1, M2, M3,
  M2, M1, M1,
  M2, M2, M3, M1,
  M2, M3, M4, M3, M1,
  M2, M3, M4, M3, M2,
  M1, M2, D5, D0,
  M3, M3, M4, M5,
  M5, M4, M3, M4, M2,
  M1, M1, M2, M3,
  M2, M1, M1
};

const float PROGMEM durt[] = {
  1, 1, 1, 1,
  1, 1, 1, 1,
  1, 1, 1, 1,
  1.5, 0.5, 2,
  1, 1, 1, 1,
  1, 1, 1, 1,
  1, 1, 1, 1,
  1.5, 0.5, 2,
  1, 1, 1, 1,
  1, 0.5, 0.5, 1, 1,
  1, 0.5, 0.5, 1, 1,
  1, 1, 1, 1,
  1, 1, 1, 1,
  1, 1, 1, 0.5, 0.5,
  1, 1, 1, 1,
  1.5, 0.5, 2,
};

const uint16_t PROGMEM birthday[] = {
  262, 262, 294, 262, 349, 330,
  262, 262, 294, 262, 392, 349,
  262, 262, 523, 440, 349, 330, 294,
  466, 466, 440, 349, 392, 349
};

const float PROGMEM birthdayDurations[] = {
  0.5, 0.5, 1, 1, 1, 2,
  0.5, 0.5, 1, 1, 1, 2,
  0.5, 0.5, 1, 1, 1, 1, 2,
  0.5, 0.5, 1, 1, 1, 2
};

const uint8_t tuneLength = sizeof(tune) / sizeof(tune[0]);
const uint8_t birthdayLength = sizeof(birthday) / sizeof(birthday[0]);

void setup() {
  Serial.begin(9600);
  
  // Initialize LCD
  mylcd.init();
  mylcd.backlight();
  mylcd.setCursor(0, 0);
  mylcd.print("password:");
  
  // Initialize servos
  servo_9.attach(9);
  servo_10.attach(10);
  servo_9.write(0);
  servo_10.write(0);
  delay(300);
  
  // Initialize pins
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
  
  Serial.println("READY");
}

void loop() {
  door();
  
  // Read sensor states
  readSensors();
  readButtons();
  
  // Send sensor data at intervals
  unsigned long currentTime = millis();
  if (currentTime - lastSensorSend >= SENSOR_INTERVAL) {
    sendSensorData();
    lastSensorSend = currentTime;
  }
  
  // Handle non-blocking music playback
  handleMusicPlayback();
  
  // Check for incoming serial commands
  if (Serial.available() > 0) {
    String receivedCommand = Serial.readString();
    receivedCommand.trim();
    
    if (receivedCommand.length() > 0) {
      processActuatorCommands(receivedCommand);
    }
  }

  delay(50);
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

void processActuatorCommands(String cmd) {
  cmd.trim();
  
  // White light control (Digital 13)
  if (cmd == "white_light_on") {
    digitalWrite(13, HIGH);
    Serial.println("ACK: White light ON");
  }
  else if (cmd == "white_light_off") {
    digitalWrite(13, LOW);
    Serial.println("ACK: White light OFF");
  }
  
  // Yellow light control (Digital 5)
  else if (cmd == "yellow_light_on") {
    digitalWrite(5, HIGH);
    Serial.println("ACK: Yellow light ON");
  }
  else if (cmd == "yellow_light_off") {
    digitalWrite(5, LOW);
    Serial.println("ACK: Yellow light OFF");
  }
  
  // Relay control (Digital 12)
  else if (cmd == "relay_on") {
    digitalWrite(12, HIGH);
    Serial.println("ACK: Relay ON");
  }
  else if (cmd == "relay_off") {
    digitalWrite(12, LOW);
    Serial.println("ACK: Relay OFF");
  }
  
  // Door servo control (Digital 9)
  else if (cmd == "door_open") {
    servo_9.write(180);
    Serial.println("ACK: Door opened");
  }
  else if (cmd == "door_close") {
    servo_9.write(0);
    Serial.println("ACK: Door closed");
  }
  else if (cmd.startsWith("door_angle=")) {
    int angle = cmd.substring(11).toInt();
    if (angle >= 0 && angle <= 180) {
      servo_9.write(angle);
      Serial.print("ACK: Door angle: ");
      Serial.println(angle);
    }
  }
  
  // Window servo control (Digital 10)
  else if (cmd == "window_open") {
    servo_10.write(180);
    Serial.println("ACK: Window opened");
  }
  else if (cmd == "window_close") {
    servo_10.write(0);
    Serial.println("ACK: Window closed");
  }
  else if (cmd.startsWith("window_angle=")) {
    int angle = cmd.substring(13).toInt();
    if (angle >= 0 && angle <= 180) {
      servo_10.write(angle);
      Serial.print("ACK: Window angle: ");
      Serial.println(angle);
    }
  }
  
  // Fan control (Digital 7 & 6)
  else if (cmd == "fan_on") {
    digitalWrite(7, LOW);
    digitalWrite(6, HIGH);
    Serial.println("ACK: Fan ON");
  }
  else if (cmd == "fan_off") {
    digitalWrite(7, LOW);
    digitalWrite(6, LOW);
    Serial.println("ACK: Fan OFF");
  }
  else if (cmd.startsWith("fan_speed=")) {
    int speed = cmd.substring(10).toInt();
    if (speed >= 0 && speed <= 255) {
      digitalWrite(7, LOW);
      analogWrite(6, speed);
      Serial.print("ACK: Fan speed: ");
      Serial.println(speed);
    }
  }
  
  // Buzzer control (Digital 3)
  else if (cmd == "buzzer_on") {
    tone(3, 1000);
    Serial.println("ACK: Buzzer ON");
  }
  else if (cmd == "buzzer_off") {
    noTone(3);
    Serial.println("ACK: Buzzer OFF");
  }
  
  // Music control
  else if (cmd == "play_birthday") {
    startBirthdaySong();
    Serial.println("ACK: Starting birthday song");
  }
  else if (cmd == "play_ode_to_joy") {
    startOdeToJoy();
    Serial.println("ACK: Starting Ode to Joy");
  }
  else if (cmd == "stop_music") {
    stopMusic();
    Serial.println("ACK: Music stopped");
  }
  else {
    Serial.print("ERROR: ");
    Serial.println(cmd);
  }
}

void startBirthdaySong() {
  isPlayingMusic = true;
  currentTuneType = 1;
  currentNote = 0;
  lastMusicNoteTime = millis();
  playCurrentNote();
}

void startOdeToJoy() {
  isPlayingMusic = true;
  currentTuneType = 2;
  currentNote = 0;
  lastMusicNoteTime = millis();
  playCurrentNote();
}

void stopMusic() {
  isPlayingMusic = false;
  noTone(tonepin);
}

void handleMusicPlayback() {
  if (!isPlayingMusic) return;
  
  unsigned long currentTime = millis();
  int noteDuration;
  uint16_t frequency;
  
  if (currentTuneType == 1) { // Birthday
    if (currentNote >= birthdayLength) {
      stopMusic();
      return;
    }
    noteDuration = (int)(pgm_read_float(&birthdayDurations[currentNote]) * 400);
    frequency = pgm_read_word(&birthday[currentNote]);
  } else { // Ode to Joy
    if (currentNote >= tuneLength) {
      stopMusic();
      return;
    }
    noteDuration = (int)(pgm_read_float(&durt[currentNote]) * 400);
    frequency = pgm_read_word(&tune[currentNote]);
  }
  
  if (currentTime - lastMusicNoteTime >= noteDuration) {
    currentNote++;
    if ((currentTuneType == 1 && currentNote < birthdayLength) ||
        (currentTuneType == 2 && currentNote < tuneLength)) {
      playCurrentNote();
      lastMusicNoteTime = currentTime;
    } else {
      stopMusic();
    }
  }
}

void playCurrentNote() {
  uint16_t frequency;
  if (currentTuneType == 1) {
    frequency = pgm_read_word(&birthday[currentNote]);
  } else {
    frequency = pgm_read_word(&tune[currentNote]);
  }
  
  if (frequency > 0) {
    tone(tonepin, frequency);
  } else {
    noTone(tonepin);
  }
}

void door() {
  static unsigned long lastBtn1Press = 0;
  static unsigned long lastBtn2Press = 0;
  static bool btn1Processed = false;
  static bool btn2Processed = false;
  
  button1 = digitalRead(4);
  button2 = digitalRead(8);

  // Button1 handling
  if (button1 == 0 && !btn1Processed) {
    if (millis() - lastBtn1Press > 50) {
      btn1_num++;
      btn1Processed = true;
      lastBtn1Press = millis();
      
      // Add to password
      if (btn1_num >= 1 && btn1_num < 5) {
        strcat(passwd, ".");
        strcat(pass, ".");
        mylcd.setCursor(0, 1);
        mylcd.print(pass);
      }
      else if (btn1_num >= 5) {
        strcat(passwd, "-");
        strcat(pass, "-");
        mylcd.setCursor(0, 1);
        mylcd.print(pass);
      }
    }
  }
  else if (button1 == 1) {
    btn1Processed = false;
    if (millis() - lastBtn1Press > 500) {
      btn1_num = 0;
    }
  }

  // Button2 handling - password check
  if (button2 == 0 && !btn2Processed) {
    if (millis() - lastBtn2Press > 50) {
      btn2Processed = true;
      lastBtn2Press = millis();
      
      if (strcmp(passwd, ".-.") == 0) {
        mylcd.clear();
        mylcd.setCursor(0, 1);
        mylcd.print("open!");
        servo_9.write(100);
      } else {
        mylcd.clear();
        mylcd.setCursor(0, 0);
        mylcd.print("error!");
        delay(1000);
        mylcd.clear();
        mylcd.setCursor(0, 0);
        mylcd.print("password:");
      }
      strcpy(passwd, "");
      strcpy(pass, "");
      btn1_num = 0;
      btn2_num = 0;
    }
  }
  else if (button2 == 1) {
    btn2Processed = false;
  }

  // Infrared sensor for auto door close
  infrar = digitalRead(2);
  if (infrar == 0) {
    servo_9.write(0);
  }
}