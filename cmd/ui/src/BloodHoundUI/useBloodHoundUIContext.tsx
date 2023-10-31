import { useContext } from 'react';
import { BloodHoundUIContext } from './BloodHoundUIContext';

const useBloodHoundUIContext = () => {
    const context = useContext(BloodHoundUIContext);
    if (context === undefined) {
        throw new Error('useBloodHoundUIContext must be used within a BloodHoundUIContextProvider');
    }
    return context;
};

export default useBloodHoundUIContext;
