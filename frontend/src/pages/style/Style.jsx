import React, { useState } from 'react';
import {BetYesButton, BetNoButton} from '../../components/buttons/trade/BetButtons';
import {ResolveButton, SelectNoButton, SelectYesButton, ConfirmResolveButton} from '../../components/buttons/marketDetails/ResolveButtons';
import SiteButton from '../../components/buttons/SiteButtons';
import SiteTabs from '../../components/tabs/SiteTabs';
import Sidebar from '../../components/sidebar/Sidebar';
import Header from '../../components/header/Header';
import {RegularInput, SuccessInput, ErrorInput, PersonInput, LockInput} from '../../components/inputs/InputBar';
import RegularInputBox from '../../components/inputs/InputBox';
import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';
import LoginModalButton from '../../components/modals/login/LoginModalClick';
import MarketsTable from '../../components/tables/MarketTables';
import MarketChart from '../../components/charts/MarketChart';
import TestMarketData from '../../tests/TestData'

const Style = () => {

    const [isSelected, setIsSelected] = useState(false);

    const tabsData = [
        { label: 'Comments', content: <div>Comments Go here...</div> },
        { label: 'Positions', content: <div>Positions Go here...</div> },
        { label: 'Bets', content: <div>Bets go here...</div> },
    ];

    return (
    <div className="overflow-auto">
    <Header />
    <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
        <thead className="bg-gray-50">
        <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Component
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Description
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Import
            </th>
        </tr>
        </thead>
        <tbody className="bg-primary-background divide-y divide-gray-200">
        <tr>
            <td className="px-6 py-4 text-white">
                <div className="flex items-center">
                    <Header />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                Header
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import Header from '../../components/header/Header';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4 ">
                <div className="flex items-center">
                    <Sidebar />
                </div>
            </td>
            <td className="px-6 py-4  text-sm text-gray-500">
                Sidebar
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import Sidebar from '../../components/sidebar/Sidebar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <BetYesButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Bet YES Button
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import BetYesButton from '../../components/buttons/trade/BetButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <BetNoButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Bet NO Button
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import BetNoButton from '../../components/buttons/BetButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <ResolveButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Neutral Button (Resolve)
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import NeutralButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <SelectNoButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Select No Button (Resolutions)
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import SelectNoButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <SelectYesButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Select Yes Button (Resolutions)
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import SelectYesButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <ConfirmResolveButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            ConfirmResolveButton (Resolutions)
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import ConfirmResolveButton from '../../components/buttons/marketDetails/ResolveButtons';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <SiteButton
                isSelected={isSelected}
                onClick={() => setIsSelected(!isSelected)}
                />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            SiteButton
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import SiteButton from '../../components/buttons/SiteButton';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <RegularInput />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                RegularInput
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { RegularInput } from '../../components/inputs/InputBar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <SuccessInput />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                SuccessInput
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { SuccessInput } from '../../components/inputs/InputBar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <ErrorInput />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                ErrorInput
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { ErrorInput } from '../../components/inputs/InputBar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <PersonInput />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                PersonInput
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { PersonInput } from '../../components/inputs/InputBar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <LockInput />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                LockInput
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { LockInput } from '../../components/inputs/InputBar';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <div className="flex items-center gap-4">
                    <RegularInputBox />
                </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                RegularInputBox
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import { RegularInputBox } from '../../components/inputs/InputBox';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
                <LoginModalButton />
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
                Login Modal Button
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>{`import LoginModalButton from '../../components/modals/LoginModalClick';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <DatetimeSelector />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Datetime Selector
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import DatetimeSelector from '../../components/datetimeSelector/DatetimeSelector';`}</code>
            </td>
        </tr>
        <tr>
            <td className="px-6 py-4">
            <div className="flex flex-wrap items-center gap-4">
                <SiteTabs tabs={tabsData} />
            </div>
            </td>
            <td className="px-6 py-4 text-sm text-gray-500">
            Tabs
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
            <code>{`import SiteTabs from '../../components/tabs/SiteTabs';`}</code>
            </td>
        </tr>
        </tbody>
    </table>
    <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
        <thead className="bg-gray-50">
        <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Description
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Import
            </th>
        </tr>
        </thead>
        <tbody className="bg-primary-background divide-y divide-gray-200">
        <tr>
            <td className="px-6 py-4 text-sm text-gray-500">
                MarketsTable
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>import MarketsTable from '../../components/tables/MarketTables';</code>
            </td>
        </tr>
        </tbody>
    </table>
    <MarketsTable ></MarketsTable>
    <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
        <thead className="bg-gray-50">
        <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Description
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Import
            </th>
        </tr>
        </thead>
        <tbody className="bg-primary-background divide-y divide-gray-200">
        <tr>
            <td className="px-6 py-4 text-sm text-gray-500">
                MarketChartDetailsTable
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>import MarketDetailsTable from '../../components/marketDetails/MarketDetailsLayout';</code>
            </td>
        </tr>
        </tbody>
    </table>
    <hr></hr>
    <p></p>
    <table className="min-w-full divide-y divide-gray-200 bg-primary-background">
        <thead className="bg-gray-50">
        <tr>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Description
            </th>
            <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
            Import
            </th>
        </tr>
        </thead>
        <tbody className="bg-primary-background divide-y divide-gray-200">
        <tr>
            <td className="px-6 py-4 text-sm text-gray-500">
                MarketChart
            </td>
            <td className="px-6 py-4 text-sm font-mono text-gray-500">
                <code>import MarketChart from '../../components/charts/MarketChart';</code>
            </td>
        </tr>
        </tbody>
    </table>
    <MarketChart
        data={TestMarketData.probabilityChanges}
        title={TestMarketData.market.questionTitle}
        className="shadow-md border border-custom-gray-light"
    />
    </div>
    );
};

export default Style;