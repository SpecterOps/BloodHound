export const downloadFile = ({ data, fileName, fileType }: { data: any; fileName: string; fileType: string }) => {
    const blob = new Blob([data], { type: fileType });
    // create an anchor tag and dispatch a click event on it to trigger download
    const a = document.createElement('a');
    a.download = fileName;
    a.href = window.URL.createObjectURL(blob);
    const clickEvent = new MouseEvent('click', {
        view: window,
        bubbles: true,
        cancelable: true,
    });
    a.dispatchEvent(clickEvent);
    a.remove();
};

export const exportToJson = (e: React.MouseEvent<Element, MouseEvent>, data: any) => {
    e.preventDefault();
    downloadFile({
        data: JSON.stringify(data),
        fileName: 'bh-graph.json',
        fileType: 'text/json',
    });
};
