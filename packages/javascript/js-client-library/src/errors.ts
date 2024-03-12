export type ErrorMessage = {
    context: string;
    message: string;
};

export type ErrorResponse = {
    http_status: number;
    timestamp: string;
    request_id: string;
    errors: ErrorMessage[];
};
