import React, { useState, useEffect } from "react";
import { gql } from "@apollo/client";
import { useQuery } from "@apollo/client/react"; // Pulls directly from the React entrypoint
import EventCard from "../components/EventCard";
import { GET_EVENTS } from "../graphql/queries";


export default function EventsDashboard() {
  const { loading, error, data } = useQuery(GET_EVENTS);
  console.log("EventsDashboard data:", data); // Log the data to see its structure
  const { events } = data || { events: [] }; // Extract events from data or provide an empty array
  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error}</div>;

  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="mx-auto max-w-7xl">
        <h2 className="mb-8 text-3xl font-bold text-gray-800">Events Dashboard</h2>

        <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-3">
          {events.length > 0 ? (
            events.map((event) => <EventCard key={event.id} event={event} />)
          ) : (
            <p>No events found</p>
          )}
        </div>
      </div>
    </div>
  );
}
