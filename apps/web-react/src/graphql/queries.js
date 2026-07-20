import { gql } from '@apollo/client';

export const GET_EVENTS = gql`
  query GetEvents {
    events {
      id
      title
      description
      bannerUrl
    }
  }
`;

export const GET_EVENT_BY_ID = gql`
  query GetEventById($id: ID!) {
    event(id: $id) {
      id
      title
      description
      date
    }
  }
`;

// Add all other queries here