// Switch to our project database context
db = db.getSiblingDB('event_platform');

// Create the collections explicitly
db.createCollection('events');
db.createCollection('bookings');

// Seed test events into the database
db.events.insertMany([
  {
    title: "Global Tech Summit 2026",
    description: "The premier event for systems architecture, Golang microservices, and high-throughput systems design.",
    slug: "global-tech-summit-2026",
    status: "active",
    date: ISODate("2026-09-15T09:00:00Z"),
    venue: {
      name: "Convention Center Hall A",
      address: "123 Innovation Way, San Francisco, CA"
    },
    inventory: {
      total_slots: 500,
      available_slots: 500
    },
    banner_url: "https://images.unsplash.com/photo-1540575467063-178a50c2df87",
    created_at: new Date()
  },
  {
    title: "React Native Deep Dive",
    description: "Build robust, native iOS and Android apps using React and share up to 90% of your business logic.",
    slug: "react-native-deep-dive",
    status: "active",
    date: ISODate("2026-10-22T13:00:00Z"),
    venue: {
      name: "Grand Ballroom B",
      address: "456 Mobile Parkway, Austin, TX"
    },
    inventory: {
      total_slots: 150,
      available_slots: 148
    },
    banner_url: "https://images.unsplash.com/photo-1591115765373-5209708f7f6f",
    created_at: new Date()
  }
]);

// Apply indexes to maintain speedy search paths
db.events.createIndex({ "slug": 1 }, { unique: true });
db.events.createIndex({ "status": 1, "date": 1 });

print("🎉 MongoDB Event Platform Database initialized and seeded successfully!");