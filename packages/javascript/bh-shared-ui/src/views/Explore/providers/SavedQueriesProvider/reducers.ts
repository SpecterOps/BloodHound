export const messageReducer = (state: any, action: any) => {
    switch (action.type) {
        case 'UPDATE':
            return { ...state, message: state.message + ' and things...' };

        default:
            return state;
    }
};
