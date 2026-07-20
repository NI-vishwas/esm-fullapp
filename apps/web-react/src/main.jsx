import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';
import { ApolloClient, HttpLink, InMemoryCache } from "@apollo/client";
import { ApolloProvider } from "@apollo/client/react";

// Initialize the client
const client = new ApolloClient({
  link: new HttpLink({ uri: "http://localhost:8080/query" }),
  cache: new InMemoryCache(),
});

ReactDOM.createRoot(document.getElementById('root')).render(
  <React.StrictMode>
    <ApolloProvider client={client}>
      <App />
    </ApolloProvider>
  </React.StrictMode>
);