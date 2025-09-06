// MongoDB initialization script
// Switch to the ultra database
db = db.getSiblingDB('ultra');

// Create the application user for the ultra database
db.createUser({
  user: 'ultra_user',
  pwd: 'ultra_password',
  roles: [
    {
      role: 'readWrite',
      db: 'ultra'
    }
  ]
});

// Create collections if they don't exist
db.createCollection('users');

// Create indexes for better performance
db.users.createIndex({ "email": 1 }, { unique: true });
db.users.createIndex({ "created_at": 1 });

print('MongoDB initialization completed for Ultra database');