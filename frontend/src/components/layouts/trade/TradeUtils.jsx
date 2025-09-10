// import API_URL from your config
import { API_URL } from '../../../config';
import React, { useState, useEffect } from 'react';

export const submitBet = (betData, token, onSuccess, onError) => {

    if (!token) {
        alert('Please log in to place a bet.');
        return;
    }

    console.log('Sending bet data:', betData);

    fetch(`${API_URL}/v0/bet`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(betData),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Success:', data);
        onSuccess(data);
    })
    .catch(error => {
        console.error('Error:', error);
        onError(error);
    });
};


export const fetchUserShares = async (marketId, token) => {
    const response = await fetch(`${API_URL}/v0/userposition/${marketId}`, {
        headers: {
        'Authorization': `Bearer ${token}` // Include the authorization token to validate user
        }
    });
    if (!response.ok) {
        throw new Error('Failed to fetch user shares');
    }
    return response.json();
    };


export const submitSale = (saleData, token, onSuccess, onError) => {
    if (!token) {
        alert('Please log in to sell shares.');
        return;
    }

    console.log('Sending sale data:', saleData);

    fetch(`${API_URL}/v0/sell`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify(saleData),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
    })
    .then(data => {
        console.log('Success:', data);
        onSuccess(data);
    })
    .catch(error => {
        console.error('Error:', error);
        onError(error);
    });
};
