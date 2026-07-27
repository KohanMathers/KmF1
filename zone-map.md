```
Root
├─ Year(int)
│  ├─ Events
│  │  └─ Event[]
│  ├─ Event(identifier)
│  │  ├─ Info
│  │  │  └─ Schedule
│  │  │     ├─ Season: int
│  │  │     ├─ RoundNumber: int
│  │  │     ├─ EventName: string
│  │  │     ├─ OfficialEventName: string
│  │  │     ├─ Country: string
│  │  │     ├─ Location: string
│  │  │     ├─ EventFormat: string
│  │  │     ├─ CircuitKey: int
│  │  │     └─ Sessions: ScheduleSession[]
│  │  │        ├─ SessionType: string
│  │  │        ├─ StartDate: timestamp?
│  │  │        ├─ StartDateUtc: timestamp
│  │  │        └─ Path: string
│  │  ├─ Sessions
│  │  │  └─ ScheduleSession[]
│  │  │     ├─ SessionType: string
│  │  │     ├─ StartDate: timestamp?
│  │  │     ├─ StartDateUtc: timestamp
│  │  │     └─ Path: string
│  │  ├─ Session(type)
│  │  │  ├─ Info
│  │  │  │  └─ SessionInfo
│  │  │  │     ├─ Meeting
│  │  │  │     │  ├─ Key: int
│  │  │  │     │  ├─ Name: string
│  │  │  │     │  ├─ OfficialName: string
│  │  │  │     │  ├─ Location: string
│  │  │  │     │  ├─ Country
│  │  │  │     │  │  ├─ Key: int
│  │  │  │     │  │  ├─ Code: string
│  │  │  │     │  │  └─ Name: string
│  │  │  │     │  └─ Circuit
│  │  │  │     │     ├─ Key: int
│  │  │  │     │     └─ ShortName: string
│  │  │  │     ├─ Key: int
│  │  │  │     ├─ Type: string
│  │  │  │     ├─ Name: string
│  │  │  │     ├─ StartDate: timestamp
│  │  │  │     ├─ EndDate: timestamp
│  │  │  │     ├─ GmtOffset: duration
│  │  │  │     ├─ Path: string
│  │  │  │     └─ ArchiveStatus
│  │  │  │        └─ Status: string
│  │  │  ├─ Drivers
│  │  │  │  └─ DriverInfo[]
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ Abbreviation: string
│  │  │  │     ├─ BroadcastName: string
│  │  │  │     ├─ FullName: string
│  │  │  │     ├─ TeamName: string
│  │  │  │     ├─ TeamColour: string
│  │  │  │     ├─ Reference: string
│  │  │  │     ├─ HeadshotUrl: string
│  │  │  │     ├─ CountryCode: string
│  │  │  │     └─ Line: int
│  │  │  ├─ Driver(identifier)
│  │  │  │  ├─ Info
│  │  │  │  │  └─ DriverInfo
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ Abbreviation: string
│  │  │  │  │     ├─ BroadcastName: string
│  │  │  │  │     ├─ FullName: string
│  │  │  │  │     ├─ TeamName: string
│  │  │  │  │     ├─ TeamColour: string
│  │  │  │  │     ├─ Reference: string
│  │  │  │  │     ├─ HeadshotUrl: string
│  │  │  │  │     ├─ CountryCode: string
│  │  │  │  │     └─ Line: int
│  │  │  │  ├─ Reference
│  │  │  │  │  └─ DriverReference
│  │  │  │  │     ├─ DriverId: string
│  │  │  │  │     ├─ Abbreviation: string
│  │  │  │  │     ├─ PermanentNumber: int?
│  │  │  │  │     ├─ GivenName: string
│  │  │  │  │     ├─ FamilyName: string
│  │  │  │  │     ├─ DateOfBirth: timestamp
│  │  │  │  │     └─ Nationality: string
│  │  │  │  ├─ Result
│  │  │  │  │  └─ SessionResults
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ DriverId: string
│  │  │  │  │     ├─ Abbreviation: string
│  │  │  │  │     ├─ TeamName: string
│  │  │  │  │     ├─ Position: int?
│  │  │  │  │     ├─ PositionText: string
│  │  │  │  │     ├─ GridPosition: int?
│  │  │  │  │     ├─ Q1: duration?
│  │  │  │  │     ├─ Q2: duration?
│  │  │  │  │     ├─ Q3: duration?
│  │  │  │  │     ├─ Laps: int?
│  │  │  │  │     ├─ Status: string
│  │  │  │  │     ├─ Points: float
│  │  │  │  │     ├─ Time: duration?
│  │  │  │  │     ├─ FastestLapRank: int?
│  │  │  │  │     ├─ FastestLapNumber: int?
│  │  │  │  │     ├─ FastestLapTime: duration?
│  │  │  │  │     └─ FastestLapAvgSpeed: float?
│  │  │  │  ├─ BestStats
│  │  │  │  │  └─ DriverBestStats
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ BestLapTime: duration?
│  │  │  │  │     ├─ BestLapNumber: int?
│  │  │  │  │     ├─ BestLapPosition: int?
│  │  │  │  │     ├─ BestSector1Time: duration?
│  │  │  │  │     ├─ BestSector1Position: int?
│  │  │  │  │     ├─ BestSector2Time: duration?
│  │  │  │  │     ├─ BestSector2Position: int?
│  │  │  │  │     ├─ BestSector3Time: duration?
│  │  │  │  │     ├─ BestSector3Position: int?
│  │  │  │  │     ├─ BestSpeedI1: int?
│  │  │  │  │     ├─ BestSpeedI1Position: int?
│  │  │  │  │     ├─ BestSpeedI2: int?
│  │  │  │  │     ├─ BestSpeedI2Position: int?
│  │  │  │  │     ├─ BestSpeedFL: int?
│  │  │  │  │     ├─ BestSpeedFLPosition: int?
│  │  │  │  │     ├─ BestSpeedST: int?
│  │  │  │  │     ├─ BestSpeedSTPosition: int?
│  │  │  │  │     └─ Timestamp: duration
│  │  │  │  ├─ Laps
│  │  │  │  │  └─ LapData[]
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ LapNo: int
│  │  │  │  │     ├─ LapTime: duration?
│  │  │  │  │     ├─ Sector1Time: duration?
│  │  │  │  │     ├─ Sector2Time: duration?
│  │  │  │  │     ├─ Sector3Time: duration?
│  │  │  │  │     ├─ SpeedI1: int?
│  │  │  │  │     ├─ SpeedI2: int?
│  │  │  │  │     ├─ SpeedFL: int?
│  │  │  │  │     ├─ SpeedST: int?
│  │  │  │  │     ├─ PitStopNo: int
│  │  │  │  │     ├─ PitInTime: duration?
│  │  │  │  │     ├─ PitOutTime: duration?
│  │  │  │  │     ├─ IsPersonalBest: bool
│  │  │  │  │     ├─ Timestamp: duration
│  │  │  │  │     ├─ LapStartTime: duration?
│  │  │  │  │     ├─ TrackStatus: string
│  │  │  │  │     ├─ Position: int?
│  │  │  │  │     ├─ FastF1Generated: bool
│  │  │  │  │     ├─ IsAccurate: bool
│  │  │  │  │     ├─ Deleted: bool
│  │  │  │  │     └─ DeletedReason: string
│  │  │  │  ├─ Lap(n)
│  │  │  │  │  └─ LapData
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ LapNo: int
│  │  │  │  │     ├─ LapTime: duration?
│  │  │  │  │     ├─ Sector1Time: duration?
│  │  │  │  │     ├─ Sector2Time: duration?
│  │  │  │  │     ├─ Sector3Time: duration?
│  │  │  │  │     ├─ SpeedI1: int?
│  │  │  │  │     ├─ SpeedI2: int?
│  │  │  │  │     ├─ SpeedFL: int?
│  │  │  │  │     ├─ SpeedST: int?
│  │  │  │  │     ├─ PitStopNo: int
│  │  │  │  │     ├─ PitInTime: duration?
│  │  │  │  │     ├─ PitOutTime: duration?
│  │  │  │  │     ├─ IsPersonalBest: bool
│  │  │  │  │     ├─ Timestamp: duration
│  │  │  │  │     ├─ LapStartTime: duration?
│  │  │  │  │     ├─ TrackStatus: string
│  │  │  │  │     ├─ Position: int?
│  │  │  │  │     ├─ FastF1Generated: bool
│  │  │  │  │     ├─ IsAccurate: bool
│  │  │  │  │     ├─ Deleted: bool
│  │  │  │  │     ├─ DeletedReason: string
│  │  │  │  │     └─ Telemetry
│  │  │  │  │        └─ Telemetry[]
│  │  │  │  │           ├─ DriverNumber: string
│  │  │  │  │           ├─ RPM: int?
│  │  │  │  │           ├─ Speed: int?
│  │  │  │  │           ├─ Gear: int?
│  │  │  │  │           ├─ Throttle: int?
│  │  │  │  │           ├─ Braking: bool?
│  │  │  │  │           ├─ DRS: string
│  │  │  │  │           ├─ DRSNo: int?
│  │  │  │  │           ├─ PositionX: int?
│  │  │  │  │           ├─ PositionY: int?
│  │  │  │  │           ├─ PositionZ: int?
│  │  │  │  │           └─ Timestamp: duration
│  │  │  │  ├─ Telemetry
│  │  │  │  │  └─ Telemetry[]
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ RPM: int?
│  │  │  │  │     ├─ Speed: int?
│  │  │  │  │     ├─ Gear: int?
│  │  │  │  │     ├─ Throttle: int?
│  │  │  │  │     ├─ Braking: bool?
│  │  │  │  │     ├─ DRS: string
│  │  │  │  │     ├─ DRSNo: int?
│  │  │  │  │     ├─ PositionX: int?
│  │  │  │  │     ├─ PositionY: int?
│  │  │  │  │     ├─ PositionZ: int?
│  │  │  │  │     └─ Timestamp: duration
│  │  │  │  ├─ Intervals
│  │  │  │  │  └─ TimingInterval[]
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ Position: int
│  │  │  │  │     ├─ GapToLeader: string
│  │  │  │  │     ├─ GapToCarAhead: string
│  │  │  │  │     └─ Timestamp: duration
│  │  │  │  ├─ Stints
│  │  │  │  │  └─ TyreStint[]
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ Compound: string
│  │  │  │  │     ├─ New: bool
│  │  │  │  │     ├─ TyresChanged: bool
│  │  │  │  │     ├─ Stint: int
│  │  │  │  │     ├─ StintAge: int
│  │  │  │  │     ├─ StartLap: int
│  │  │  │  │     ├─ LapFlags: int
│  │  │  │  │     └─ Timestamp: duration
│  │  │  │  ├─ PitStops
│  │  │  │  │  └─ PitStop[]
│  │  │  │  │     ├─ DriverId: string
│  │  │  │  │     ├─ StopNumber: int
│  │  │  │  │     ├─ Lap: int
│  │  │  │  │     ├─ Duration: duration
│  │  │  │  │     └─ Time: duration
│  │  │  │  └─ TeamRadio
│  │  │  │     └─ TeamRadio[]
│  │  │  │        ├─ DriverNumber: string
│  │  │  │        ├─ Path: string
│  │  │  │        ├─ Utc: timestamp
│  │  │  │        └─ Timestamp: duration
│  │  │  ├─ Laps
│  │  │  │  └─ LapData[]
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ LapNo: int
│  │  │  │     ├─ LapTime: duration?
│  │  │  │     ├─ Sector1Time: duration?
│  │  │  │     ├─ Sector2Time: duration?
│  │  │  │     ├─ Sector3Time: duration?
│  │  │  │     ├─ SpeedI1: int?
│  │  │  │     ├─ SpeedI2: int?
│  │  │  │     ├─ SpeedFL: int?
│  │  │  │     ├─ SpeedST: int?
│  │  │  │     ├─ PitStopNo: int
│  │  │  │     ├─ PitInTime: duration?
│  │  │  │     ├─ PitOutTime: duration?
│  │  │  │     ├─ IsPersonalBest: bool
│  │  │  │     ├─ Timestamp: duration
│  │  │  │     ├─ LapStartTime: duration?
│  │  │  │     ├─ TrackStatus: string
│  │  │  │     ├─ Position: int?
│  │  │  │     ├─ FastF1Generated: bool
│  │  │  │     ├─ IsAccurate: bool
│  │  │  │     ├─ Deleted: bool
│  │  │  │     └─ DeletedReason: string
│  │  │  ├─ Results
│  │  │  │  └─ SessionResults[]
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ DriverId: string
│  │  │  │     ├─ Abbreviation: string
│  │  │  │     ├─ TeamName: string
│  │  │  │     ├─ Position: int?
│  │  │  │     ├─ PositionText: string
│  │  │  │     ├─ GridPosition: int?
│  │  │  │     ├─ Q1: duration?
│  │  │  │     ├─ Q2: duration?
│  │  │  │     ├─ Q3: duration?
│  │  │  │     ├─ Laps: int?
│  │  │  │     ├─ Status: string
│  │  │  │     ├─ Points: float
│  │  │  │     ├─ Time: duration?
│  │  │  │     ├─ FastestLapRank: int?
│  │  │  │     ├─ FastestLapNumber: int?
│  │  │  │     ├─ FastestLapTime: duration?
│  │  │  │     └─ FastestLapAvgSpeed: float?
│  │  │  ├─ BestStats
│  │  │  │  └─ DriverBestStats[]
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ BestLapTime: duration?
│  │  │  │     ├─ BestLapNumber: int?
│  │  │  │     ├─ BestLapPosition: int?
│  │  │  │     ├─ BestSector1Time: duration?
│  │  │  │     ├─ BestSector1Position: int?
│  │  │  │     ├─ BestSector2Time: duration?
│  │  │  │     ├─ BestSector2Position: int?
│  │  │  │     ├─ BestSector3Time: duration?
│  │  │  │     ├─ BestSector3Position: int?
│  │  │  │     ├─ BestSpeedI1: int?
│  │  │  │     ├─ BestSpeedI1Position: int?
│  │  │  │     ├─ BestSpeedI2: int?
│  │  │  │     ├─ BestSpeedI2Position: int?
│  │  │  │     ├─ BestSpeedFL: int?
│  │  │  │     ├─ BestSpeedFLPosition: int?
│  │  │  │     ├─ BestSpeedST: int?
│  │  │  │     ├─ BestSpeedSTPosition: int?
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ Weather
│  │  │  │  └─ WeatherSample[]
│  │  │  │     ├─ AirTemp: float
│  │  │  │     ├─ TrackTemp: float
│  │  │  │     ├─ Humidity: float
│  │  │  │     ├─ Pressure: float
│  │  │  │     ├─ Rainfall: bool
│  │  │  │     ├─ WindDirection: int
│  │  │  │     ├─ WindSpeed: float
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ TrackStatus
│  │  │  │  └─ TrackStatus[]
│  │  │  │     ├─ Status: string
│  │  │  │     ├─ Message: string
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ Progress
│  │  │  │  └─ SessionProgress[]
│  │  │  │     ├─ CurrentLap: int
│  │  │  │     ├─ TotalLaps: int
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ StatusEvents
│  │  │  │  └─ SessionStatusEvent[]
│  │  │  │     ├─ Status: string
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ Clock
│  │  │  │  └─ SessionClock[]
│  │  │  │     ├─ Remaining: duration
│  │  │  │     ├─ Extrapolating: bool
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ ArchiveStatus
│  │  │  │  └─ ArchiveStatus
│  │  │  │     └─ Status: string
│  │  │  ├─ RaceControlMessages
│  │  │  │  └─ RaceControlMessage[]
│  │  │  │     ├─ Message: string
│  │  │  │     ├─ Category: string
│  │  │  │     ├─ Flag: string
│  │  │  │     ├─ Scope: string
│  │  │  │     ├─ Sector: int?
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ Lap: int
│  │  │  │     ├─ Status: string
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ TeamRadio
│  │  │  │  └─ TeamRadio[]
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ Path: string
│  │  │  │     ├─ Utc: timestamp
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ PitStops
│  │  │  │  └─ PitStop[]
│  │  │  │     ├─ DriverId: string
│  │  │  │     ├─ StopNumber: int
│  │  │  │     ├─ Lap: int
│  │  │  │     ├─ Duration: duration
│  │  │  │     └─ Time: duration
│  │  │  ├─ TopThree
│  │  │  │  └─ TopThreeLive[]
│  │  │  │     ├─ Position: int
│  │  │  │     ├─ DriverNumber: string
│  │  │  │     ├─ Abbreviation: string
│  │  │  │     ├─ BroadcastName: string
│  │  │  │     ├─ FullName: string
│  │  │  │     ├─ TeamName: string
│  │  │  │     ├─ TeamColour: string
│  │  │  │     ├─ LapTime: duration?
│  │  │  │     ├─ LapState: int
│  │  │  │     ├─ DiffToAhead: string
│  │  │  │     ├─ DiffToLeader: string
│  │  │  │     ├─ OverallFastest: bool
│  │  │  │     ├─ PersonalFastest: bool
│  │  │  │     └─ Timestamp: duration
│  │  │  ├─ ChampionshipPrediction
│  │  │  │  ├─ Drivers
│  │  │  │  │  └─ ChampionshipPredictionDriver[]
│  │  │  │  │     ├─ DriverNumber: string
│  │  │  │  │     ├─ CurrentPosition: int
│  │  │  │  │     ├─ CurrentPoints: float
│  │  │  │  │     ├─ PredictedPosition: int
│  │  │  │  │     ├─ PredictedPoints: float
│  │  │  │  │     └─ Timestamp: duration
│  │  │  │  └─ Teams
│  │  │  │     └─ ChampionshipPredictionTeam[]
│  │  │  │        ├─ TeamId: string
│  │  │  │        ├─ TeamName: string
│  │  │  │        ├─ CurrentPosition: int
│  │  │  │        ├─ CurrentPoints: float
│  │  │  │        ├─ PredictedPosition: int
│  │  │  │        ├─ PredictedPoints: float
│  │  │  │        └─ Timestamp: duration
│  │  │  └─ Broadcast
│  │  │     ├─ Audio
│  │  │     │  └─ AudioStream[]
│  │  │     │     ├─ Name: string
│  │  │     │     ├─ Language: string
│  │  │     │     ├─ Uri: string
│  │  │     │     ├─ Path: string
│  │  │     │     └─ Utc: timestamp?
│  │  │     └─ Content
│  │  │        └─ ContentStream[]
│  │  │           ├─ Type: string
│  │  │           ├─ Name: string
│  │  │           ├─ Language: string
│  │  │           ├─ Uri: string
│  │  │           ├─ Path: string
│  │  │           └─ Utc: timestamp?
│  │  ├─ Circuit
│  │  │  ├─ Corners
│  │  │  │  └─ Corner[]
│  │  │  │     ├─ Number: int
│  │  │  │     ├─ Letter: string
│  │  │  │     ├─ X: float
│  │  │  │     ├─ Y: float
│  │  │  │     ├─ Angle: float
│  │  │  │     └─ Length: float
│  │  │  ├─ MarshalLights
│  │  │  │  └─ MarshalLight[]
│  │  │  │     ├─ Number: int
│  │  │  │     ├─ Letter: string
│  │  │  │     ├─ X: float
│  │  │  │     ├─ Y: float
│  │  │  │     ├─ Angle: float
│  │  │  │     └─ Length: float
│  │  │  └─ MarshalSectors
│  │  │     └─ MarshalSector[]
│  │  │        ├─ Number: int
│  │  │        ├─ Letter: string
│  │  │        ├─ X: float
│  │  │        ├─ Y: float
│  │  │        ├─ Angle: float
│  │  │        └─ Length: float
│  │  └─ Standings
│  │     ├─ Drivers
│  │     │  └─ DriverStanding[]
│  │     │     ├─ Season: int
│  │     │     ├─ Round: int
│  │     │     ├─ Position: int
│  │     │     ├─ PositionText: string
│  │     │     ├─ Points: float
│  │     │     ├─ Wins: int
│  │     │     ├─ DriverId: string
│  │     │     ├─ Abbreviation: string
│  │     │     └─ ConstructorIds: string[]
│  │     └─ Constructors
│  │        └─ ConstructorStanding[]
│  │           ├─ Season: int
│  │           ├─ Round: int
│  │           ├─ Position: int
│  │           ├─ PositionText: string
│  │           ├─ Points: float
│  │           ├─ Wins: int
│  │           └─ ConstructorId: string
│  ├─ TeamStyle(identifier)
│  │  └─ TeamStyle
│  │     ├─ Key: string
│  │     ├─ ShortName: string
│  │     └─ Colors
│  │        ├─ Official: string
│  │        └─ FastF1: string
│  ├─ TeamStyles
│  │  └─ TeamStyle[]
│  │     ├─ Key: string
│  │     ├─ ShortName: string
│  │     └─ Colors
│  │        ├─ Official: string
│  │        └─ FastF1: string
│  ├─ CompoundColor(compound): string
│  └─ Compounds: string[]
├─ DriverReferences
│  └─ DriverReference[]
│     ├─ DriverId: string
│     ├─ Abbreviation: string
│     ├─ PermanentNumber: int?
│     ├─ GivenName: string
│     ├─ FamilyName: string
│     ├─ DateOfBirth: timestamp
│     └─ Nationality: string
├─ DriverReference(id)
│  └─ DriverReference
│     ├─ DriverId: string
│     ├─ Abbreviation: string
│     ├─ PermanentNumber: int?
│     ├─ GivenName: string
│     ├─ FamilyName: string
│     ├─ DateOfBirth: timestamp
│     └─ Nationality: string
├─ Constructors
│  └─ ConstructorReference[]
│     ├─ ConstructorId: string
│     ├─ Name: string
│     └─ Nationality: string
├─ Constructor(id)
│  └─ ConstructorReference
│     ├─ ConstructorId: string
│     ├─ Name: string
│     └─ Nationality: string
├─ CircuitReferences
│  └─ CircuitReference[]
│     ├─ CircuitId: string
│     ├─ Name: string
│     ├─ Locality: string
│     ├─ Country: string
│     ├─ Latitude: float
│     └─ Longitude: float
├─ CircuitReference(id)
│  └─ CircuitReference
│     ├─ CircuitId: string
│     ├─ Name: string
│     ├─ Locality: string
│     ├─ Country: string
│     ├─ Latitude: float
│     └─ Longitude: float
└─ FinishingStatusReferences
   └─ FinishingStatusReference[]
      ├─ StatusId: string
      ├─ Status: string
      └─ Count: int
```
