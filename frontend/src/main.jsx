import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App.jsx';
import './index.css';
import { createHTMLDocument } from 'https://deno.land/std@0.165.0/node/html.js';


if (!!!document) {
  document = createHTMLDocument();
}

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);
