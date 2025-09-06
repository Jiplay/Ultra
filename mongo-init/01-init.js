// MongoDB initialization script
db = db.getSiblingDB('ultra');

// Create collections if they don't exist
db.createCollection('users');

// Create indexes for better performance
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "created_at": 1 });

print('MongoDB initialization completed for Ultra database');