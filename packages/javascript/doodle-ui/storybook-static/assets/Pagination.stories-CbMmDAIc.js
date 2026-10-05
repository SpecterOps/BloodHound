import{j as i}from"./jsx-runtime-BTJTZTIL.js";import{r as C}from"./index-2AqrBsEA.js";import{P as m}from"./Pagination-CXTveDON.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index.es-C0EmWI2V.js";import"./index-TwK_z8r1.js";import"./Select-g8-1NtLo.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-Dv0R_RYs.js";import"./index-D2ZHEr6j.js";import"./index-Dc_EP5I0.js";import"./index-DnVAETpg.js";import"./component-D-Cd_cec.js";import"./index-BQHA1h7N.js";import"./index-ejpvWIAw.js";import"./index-DSMk1t8m.js";import"./floating-ui.react-dom-D40P0bPq.js";import"./index-DRLFypRQ.js";import"./index-CFknkR0k.js";import"./index-oIeaYTeP.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";const I={title:"Components/Pagination",component:m,parameters:{layout:"center"},tags:["autodocs"],argTypes:{},args:{}},l=r=>{const[g,n]=C.useState({...r}),P=t=>{n(e=>({...e,page:t}))},c=t=>{const e=t??10;n(u=>({...u,rowsPerPage:e}))};return i.jsx(m,{...g,onPageChange:P,onRowsPerPageChange:c})},o={args:{page:0,rowsPerPage:10,count:123,onPageChange(){},onRowsPerPageChange(){}},render:r=>i.jsx(l,{...r})};var a,s,p;o.parameters={...o.parameters,docs:{...(a=o.parameters)==null?void 0:a.docs,source:{originalSource:`{
  args: {
    page: 0,
    rowsPerPage: 10,
    count: 123,
    onPageChange() {},
    onRowsPerPageChange() {}
  },
  render: props => {
    return <PaginationController {...props} />;
  }
}`,...(p=(s=o.parameters)==null?void 0:s.docs)==null?void 0:p.source}}};const J=["PaginationControls"];export{o as PaginationControls,J as __namedExportsOrder,I as default};
