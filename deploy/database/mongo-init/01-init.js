// Ultra API MongoDB Initialization Script
// This script creates collections and indexes for the Ultra application

// Switch to the ultra database
db = db.getSiblingDB('ultra');

// Create users collection with indexes
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "created_at": 1 });

// Create sessions collection with indexes
db.sessions.createIndex({ "user_id": 1 });
db.sessions.createIndex({ "expires_at": 1 }, { expireAfterSeconds: 0 });
db.sessions.createIndex({ "token": 1 }, { unique: true });

// Create logs collection with indexes
db.logs.createIndex({ "timestamp": 1 });
db.logs.createIndex({ "level": 1 });
db.logs.createIndex({ "user_id": 1 });

// Create analytics collection with indexes
db.analytics.createIndex({ "user_id": 1 });
db.analytics.createIndex({ "event_type": 1 });
db.analytics.createIndex({ "timestamp": 1 });

// Insert sample user data
db.users.insertMany([
    {
        _id: ObjectId(),
        email: "demo@ultra.com",
        name: "Demo User",
        password_hash: "$2a$10$rQ3.example.hash", // This would be a real bcrypt hash
        created_at: new Date(),
        updated_at: new Date(),
        is_active: true,
        preferences: {
            theme: "light",
            notifications: true,
            units: "metric"
        }
    },
    {
        _id: ObjectId(),
        email: "admin@ultra.com",
        name: "Admin User",
        password_hash: "$2a$10$rQ3.example.hash2",
        created_at: new Date(),
        updated_at: new Date(),
        is_active: true,
        role: "admin",
        preferences: {
            theme: "dark",
            notifications: true,
            units: "imperial"
        }
    }
]);

// Insert sample analytics events
db.analytics.insertMany([
    {
        user_id: "demo-user",
        event_type: "login",
        timestamp: new Date(),
        data: {
            ip: "127.0.0.1",
            user_agent: "Ultra Mobile App"
        }
    },
    {
        user_id: "demo-user",
        event_type: "program_created",
        timestamp: new Date(),
        data: {
            program_name: "Full Body Strength",
            workout_count: 3
        }
    },
    {
        user_id: "demo-user",
        event_type: "recipe_created",
        timestamp: new Date(),
        data: {
            recipe_name: "Protein Power Bowl",
            servings: 2
        }
    }
]);

print("MongoDB initialization completed successfully!");
print("Collections created: users, sessions, logs, analytics");
print("Sample data inserted for development");

// List all collections to verify
print("Available collections:");
db.getCollectionNames().forEach(function(collection) {
    print("- " + collection);
});