import React, { useState, useEffect } from 'react';
import { API_URL } from '../../../../config';
import LoadingSpinner from '../../../loaders/LoadingSpinner';

const UserFinancialStatementsLayout = ({ username }) => {
    const [financialData, setFinancialData] = useState(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);

    useEffect(() => {
        const fetchFinancialData = async () => {
            try {
                console.log(`Fetching financial data for user: ${username} from ${API_URL}/v0/users/${username}/financial`);
                const response = await fetch(`${API_URL}/v0/users/${username}/financial`);
                if (response.ok) {
                    const data = await response.json();
                    console.log('Financial data:', data);
                    setFinancialData(data.financial);
                } else {
                    throw new Error(`Error fetching financial data: ${response.statusText}`);
                }
            } catch (err) {
                console.error('Error fetching financial data:', err);
                setError(err.message);
            } finally {
                setLoading(false);
            }
        };

        if (username) {
            fetchFinancialData();
        }
    }, [username]);

    const formatValue = (value, isAmount = true) => {
        if (value === null || value === undefined) return 'N/A';
        if (isAmount) {
            const colorClass = value >= 0 ? 'text-green-400' : 'text-red-400';
            return <span className={colorClass}>{value.toLocaleString()}</span>;
        }
        return value.toLocaleString();
    };

    const FinancialSection = ({ title, items, bgColor = 'bg-gray-800' }) => (
        <div className="mb-8">
            <h4 className={`text-lg font-semibold text-white mb-4 p-3 ${bgColor} rounded-t-lg`}>
                {title}
            </h4>
            <table className="w-full divide-y divide-gray-200 bg-primary-background rounded-b-lg overflow-hidden">
                <thead className="bg-gray-50">
                    <tr>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Metric
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Value
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Description
                        </th>
                        <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Formula
                        </th>
                    </tr>
                </thead>
                <tbody className="bg-primary-background divide-y divide-gray-200">
                    {items.map((item, index) => (
                        <tr key={index} className="hover:bg-gray-700">
                            <td className="px-4 py-4 text-sm font-medium text-white">
                                {item.name}
                            </td>
                            <td className="px-4 py-4 text-sm font-mono">
                                {formatValue(item.value)}
                            </td>
                            <td className="px-4 py-4 text-sm text-gray-300">
                                {item.description}
                            </td>
                            <td className="px-4 py-4 text-sm font-mono text-gray-400">
                                {item.formula}
                            </td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );

    if (loading) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="flex items-center justify-center">
                    <LoadingSpinner />
                    <span className="ml-2 text-gray-300">Loading financial data...</span>
                </div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-red-400">
                    Error loading financial data: {error}
                </div>
            </div>
        );
    }

    if (!financialData) {
        return (
            <div className="bg-primary-background shadow-md rounded-lg p-6">
                <div className="text-gray-300">
                    No financial data available.
                </div>
            </div>
        );
    }

    // Balance Sheet Data
    const balanceSheetItems = [
        {
            name: 'Account Balance',
            value: financialData.accountBalance,
            description: 'Current available funds (cash equivalent)',
            formula: 'Initial Balance + All Gains/Losses'
        },
        {
            name: 'Amount In Play',
            value: financialData.amountInPlay,
            description: 'Total current value of all positions',
            formula: 'Sum of all position values'
        },
        {
            name: 'Amount Borrowed',
            value: financialData.amountBorrowed,
            description: 'Amount owed when account balance is negative',
            formula: 'max(0, -accountBalance)'
        },
        {
            name: 'Retained Earnings',
            value: financialData.retainedEarnings,
            description: 'Funds not currently invested in positions',
            formula: 'accountBalance - amountInPlay'
        },
        {
            name: 'Total Equity',
            value: financialData.equity,
            description: 'Total financial position after liabilities',
            formula: 'retainedEarnings + amountInPlay - amountBorrowed'
        },
        {
            name: 'Maximum Debt Allowed',
            value: financialData.maximumDebtAllowed,
            description: 'Credit limit for borrowing',
            formula: 'System configuration limit'
        }
    ];

    // Income Statement Data
    const incomeStatementItems = [
        {
            name: 'Trading Profits',
            value: financialData.tradingProfits,
            description: 'Net gains/losses from market positions',
            formula: 'Sum of (position.value - position.totalSpent)'
        },
        {
            name: 'Work Profits',
            value: financialData.workProfits,
            description: 'Earnings from work rewards and bounties',
            formula: 'Sum of WorkReward + Bounty transactions'
        },
        {
            name: 'Total Profits',
            value: financialData.totalProfits,
            description: 'Combined profits from all sources',
            formula: 'tradingProfits + workProfits'
        },
        {
            name: 'Realized Profits',
            value: financialData.realizedProfits,
            description: 'Confirmed gains/losses from resolved positions',
            formula: 'Profits from resolved markets only'
        },
        {
            name: 'Potential Profits',
            value: financialData.potentialProfits,
            description: 'Unrealized gains/losses from active positions',
            formula: 'Profits from unresolved markets only'
        }
    ];

    // Cash Flow Data
    const cashFlowItems = [
        {
            name: 'Total Spent',
            value: financialData.totalSpent,
            description: 'All money invested across all markets',
            formula: 'Sum of all bet amounts ever placed'
        },
        {
            name: 'Total Spent In Play',
            value: financialData.totalSpentInPlay,
            description: 'Money invested in current positions',
            formula: 'Sum of totalSpent for current positions'
        },
        {
            name: 'Amount In Play (Active)',
            value: financialData.amountInPlayActive,
            description: 'Current value of unresolved positions only',
            formula: 'Sum of position values (unresolved markets)'
        }
    ];

    // Market Position Data
    const marketPositionItems = [
        {
            name: 'Realized Value',
            value: financialData.realizedValue,
            description: 'Final value from resolved positions',
            formula: 'Sum of values from resolved markets'
        },
        {
            name: 'Potential Value',
            value: financialData.potentialValue,
            description: 'Current estimated value of active positions',
            formula: 'Sum of values from unresolved markets'
        },
        {
            name: 'Position Efficiency',
            value: financialData.totalSpentInPlay > 0 ?
                Math.round((financialData.amountInPlayActive / financialData.totalSpentInPlay) * 100) : 0,
            description: 'Current position value vs amount spent (%)',
            formula: '(amountInPlayActive / totalSpentInPlay) × 100'
        }
    ];

    return (
        <div className="bg-primary-background shadow-md rounded-lg">
            <div className="p-6">
                <h3 className="text-xl font-bold text-white mb-6">Financial Statements</h3>

                <FinancialSection
                    title="Balance Sheet - Financial Position"
                    items={balanceSheetItems}
                    bgColor="bg-info-blue"
                />

                <FinancialSection
                    title="Income Statement - Profitability"
                    items={incomeStatementItems}
                    bgColor="bg-green-btn"
                />

                <FinancialSection
                    title="Cash Flow Statement - Investment Activity"
                    items={cashFlowItems}
                    bgColor="bg-warning-orange"
                />

                <FinancialSection
                    title="Market Position Summary - Trading Performance"
                    items={marketPositionItems}
                    bgColor="bg-primary-pink"
                />
            </div>
        </div>
    );
};

export default UserFinancialStatementsLayout;
