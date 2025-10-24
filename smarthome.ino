  #include <Servo.h>
  #include <Wire.h>
  #include <LiquidCrystal_I2C.h>

  // Set the communication address of I2C to 0x27, display 16 characters every line, two lines in total
  LiquidCrystal_I2C mylcd(0x27, 16, 2);

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
  volatile int btn1_num = 0; // Counter for button1 presses
  volatile int btn2_num = 0; // Counter for button2 presses

  // Password variables
  String pass = "";     // Display string for LCD
  String passwd = "";   // Actual password string
  volatile int flag = 0, flag2 = 0, flag3 = 0; // Sensor flags

  // Control variables
  String command;
  volatile int val; // For single character commands

  // Music variables
  int length;
  int tonepin = 3; // Set the signal end of passive buzzer to digital 3

  // Define sound frequencies
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

  // Ode to Joy tune
  int tune[] = {
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

  // Music beat
  float durt[] = {
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

  void setup() {
    Serial.begin(9600);
      // Initialize LCD
    mylcd.init();
    mylcd.backlight();
    // LCD shows "password:" at first row and column
    mylcd.setCursor(0, 0);
    mylcd.print("password:");
    
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
    
    length = sizeof(tune) / sizeof(tune[0]); // Set the value of length
  }
  void loop() {
    door();
    
    // Read sensor states for serial output
    readSensors();
    readButtons();
    sendSensorData();
    
    // Check for incoming serial commands
    if (Serial.available() > 0) {
      command = Serial.readStringUntil('\n');
      command.trim();
      if (command.length() > 0) {
        processActuatorCommands();
        Serial.println("ACK");
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
    // Music control
    else if (command.startsWith("play_birthday")) {
      music1();
      Serial.println("Playing birthday song");
    }
    else if (command.startsWith("play_ode_to_joy")) {
      music2();
      Serial.println("Playing Ode to Joy");
    }
    else if (command.startsWith("stop_music")) {
      noTone(3);
      Serial.println("Music stopped");
    }
    else {
      Serial.print("Unknown command: ");
      Serial.println(command);
    }
  }  void door() {
    static unsigned long lastBtn1Press = 0;
    static unsigned long lastBtn2Press = 0;
    static bool btn1Processed = false;
    static bool btn2Processed = false;
    
    button1 = digitalRead(4);
    button2 = digitalRead(8);

    // Button1 handling - non-blocking
    if (button1 == 0 && !btn1Processed) {
      if (millis() - lastBtn1Press > 50) { // Debounce
        btn1_num++;
        btn1Processed = true;
        lastBtn1Press = millis();
        
        // Add to password based on press count
        if (btn1_num >= 1 && btn1_num < 5) {
          Serial.print(".");
          passwd = String(passwd) + String(".");
          pass = String(pass) + String(".");
          mylcd.setCursor(0, 1);
          mylcd.print(pass);
        }
        else if (btn1_num >= 5) {
          Serial.print("-");
          passwd = String(passwd) + String("-");
          pass = String(pass) + String("-");
          mylcd.setCursor(0, 1);
          mylcd.print(pass);
        }
      }
    }
    else if (button1 == 1) {
      btn1Processed = false;
      if (millis() - lastBtn1Press > 500) { // Reset counter after 500ms
        btn1_num = 0;
      }
    }

    // Button2 handling - password check
    if (button2 == 0 && !btn2Processed) {
      if (millis() - lastBtn2Press > 50) { // Debounce
        btn2Processed = true;
        lastBtn2Press = millis();
        
        if (passwd == ".--.-." || passwd == ".--.-.") {
          mylcd.clear();
          mylcd.setCursor(0, 1);
          mylcd.print("open!");
          servo_9.write(100);
          delay(300);
          // Use non-blocking delay for door open time
          // Door will be closed by infrared sensor
        } else {
          mylcd.clear();
          mylcd.setCursor(0, 0);
          mylcd.print("error!");
          delay(1000);
          mylcd.clear();
          mylcd.setCursor(0, 0);
          mylcd.print("password:");
        }
        passwd = "";
        pass = "";
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

  // Birthday song function
  void music1() {
    int birthday[] = {262, 262, 294, 262, 349, 330,
                      262, 262, 294, 262, 392, 349,
                      262, 262, 523, 440, 349, 330, 294,
                      466, 466, 440, 349, 392, 349};
    
    float birthdayDurations[] = {0.5, 0.5, 1, 1, 1, 2,
                                0.5, 0.5, 1, 1, 1, 2,
                                0.5, 0.5, 1, 1, 1, 1, 2,
                                0.5, 0.5, 1, 1, 1, 2};
    
    int birthdayLength = sizeof(birthday) / sizeof(birthday[0]);
    
    for (int i = 0; i < birthdayLength; i++) {
      if (birthday[i] != 0) {
        tone(tonepin, birthday[i]);
        delay(int(birthdayDurations[i] * 400));
        noTone(tonepin);
        delay(50);
      } else {
        delay(int(birthdayDurations[i] * 400));
      }
    }
  }

  // Ode to Joy function
  void music2() {
    for (int x = 0; x < length; x++) {
      if (tune[x] != 0) {
        tone(tonepin, tune[x]);
        delay(int(durt[x] * 400));
        noTone(tonepin);
        delay(50);
      } else {
        delay(int(durt[x] * 400));
      }
    }
  }
