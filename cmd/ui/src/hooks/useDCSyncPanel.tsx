// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

import { createContext, useContext, useState, useEffect } from 'react';
import { useExploreParams, useExploreGraph } from 'bh-shared-ui';

interface DCSyncPanelContextType {
    showDCSyncPanel: boolean;
    setShowDCSyncPanel: (show: boolean) => void;
    isDCSyncSearch: boolean;
    setIsDCSyncSearch: (isDCSync: boolean) => void;
}

const DCSyncPanelContext = createContext<DCSyncPanelContextType | undefined>(undefined);

export const DCSyncPanelProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [isDCSyncSearch, setIsDCSyncSearch] = useState(false);
    const [showDCSyncPanel, setShowDCSyncPanel] = useState(false);
    
    // FOR TESTING: Force show DCSync panel
    const [forceShowForTesting] = useState(true);
    
    const { exploreSearchTab } = useExploreParams();
    const exploreGraphResult = useExploreGraph();
    const graphData = exploreGraphResult.data;

    // Track if the explore search tab is 'sniffdeep' to help identify DCSync context
    useEffect(() => {
        console.log('useDCSyncPanel: exploreSearchTab changed to:', exploreSearchTab);
        if (exploreSearchTab !== 'sniffdeep' && isDCSyncSearch) {
            console.log('useDCSyncPanel: Resetting isDCSyncSearch due to tab change');
            setIsDCSyncSearch(false);
        }
    }, [exploreSearchTab, isDCSyncSearch]);

    useEffect(() => {
        console.log('useDCSyncPanel: isDCSyncSearch changed to:', isDCSyncSearch);
        console.log('useDCSyncPanel: graphData available:', !!graphData);
        console.log('useDCSyncPanel: forceShowForTesting:', forceShowForTesting);
        
        if (forceShowForTesting) {
            console.log('useDCSyncPanel: TESTING MODE - Forcing panel to show');
            setShowDCSyncPanel(true);
        } else if (isDCSyncSearch) {
            console.log('useDCSyncPanel: Setting showDCSyncPanel to true');
            setShowDCSyncPanel(true);
        } else {
            console.log('useDCSyncPanel: Setting showDCSyncPanel to false');
            setShowDCSyncPanel(false);
        }
    }, [isDCSyncSearch, graphData, forceShowForTesting]);

    console.log('useDCSyncPanel: Current state - isDCSyncSearch:', isDCSyncSearch, 'showDCSyncPanel:', showDCSyncPanel);

    const value: DCSyncPanelContextType = {
        isDCSyncSearch,
        setIsDCSyncSearch,
        showDCSyncPanel,
        setShowDCSyncPanel,
    };

    return <DCSyncPanelContext.Provider value={value}>{children}</DCSyncPanelContext.Provider>;
};

export const useDCSyncPanel = () => {
    const context = useContext(DCSyncPanelContext);
    if (context === undefined) {
        throw new Error('useDCSyncPanel must be used within a DCSyncPanelProvider');
    }
    return context;
};
