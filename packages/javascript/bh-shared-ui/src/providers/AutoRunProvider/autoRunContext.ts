import { createContext } from 'react';

const AutoRunContext = createContext({
    autoRun: true,
    setAutoRun: (autoRunQueries: boolean) => {},
});

export default AutoRunContext;
