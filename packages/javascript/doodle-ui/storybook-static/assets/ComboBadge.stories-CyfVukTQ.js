import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{F as C,o as F}from"./index.es-C0EmWI2V.js";import{R as B}from"./RadialGauge-DNUeI5B2.js";import{c as w}from"./index-Bb4qSo10.js";import{c as i}from"./utils-BIldDEfN.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";const J=w("relative leading-[1] z-10 h-8 text-nowrap",{variants:{type:{slideLeft:"clip-right-rounded flex flex-row-reverse",slideRight:"clip-left-rounded",inlineSlideLeft:"clip-right-rounded flex items-center flex-row flex-row-reverse",inlineSlideRight:"clip-left-rounded flex items-center flex-row"}}}),X=w("bg-primary text-neutral-light-1 rounded px-4 py-2 z-10 transition-[transform] duration-300",{variants:{type:{slideLeft:"absolute left-2 bottom-0 top-0 peer-hover:-translate-x-full",slideRight:"absolute right-2 bottom-0 top-0 peer-hover:translate-x-full",inlineSlideLeft:"inline-block peer-hover:translate-x-0 translate-x-full -mr-2",inlineSlideRight:"inline-block peer-hover:translate-x-0 -translate-x-full -ml-2"},displayAdornment:{true:"translate-x-0 peer-hover:translate-x-0"}},compoundVariants:[{type:["slideLeft"],displayAdornment:!0,className:"inline-block relative left-0 -mr-2"},{type:["slideRight"],displayAdornment:!0,className:"inline-block relative right-0 -ml-2"}]}),o=l=>{const{label:j,adornment:d,ariaLabel:E,type:m="slideLeft",displayAdornment:T=!1,disableAdornment:k=!1,className:q,...I}=l;return e.jsxs("div",{role:"status","aria-label":E,className:i(J({type:m}),q),...I,children:[e.jsx("div",{className:i("bg-neutral-light-5 text-primary dark:bg-neutral-dark-5 dark:text-neutral-light-1 rounded inline-block px-4 py-2 relative z-20",{peer:!k}),children:j}),d?e.jsx("div",{className:i(X({type:m,displayAdornment:T})),children:d}):null]})};o.__docgenInfo={description:"",methods:[],displayName:"ComboBadge",props:{label:{required:!0,tsType:{name:"union",raw:"string | JSX.Element",elements:[{name:"string"},{name:"JSX.Element"}]},description:""},adornment:{required:!1,tsType:{name:"union",raw:"string | JSX.Element",elements:[{name:"string"},{name:"JSX.Element"}]},description:""},ariaLabel:{required:!0,tsType:{name:"string"},description:"An arial label is required because of the hidden adornment. Please provide context as to what this badge is conveying to a visual user"},type:{required:!1,tsType:{name:"union",raw:"'slideLeft' | 'slideRight' | 'inlineSlideLeft' | 'inlineSlideRight'",elements:[{name:"literal",value:"'slideLeft'"},{name:"literal",value:"'slideRight'"},{name:"literal",value:"'inlineSlideLeft'"},{name:"literal",value:"'inlineSlideRight'"}]},description:"slideLeft by default"},displayAdornment:{required:!1,tsType:{name:"boolean"},description:"false by default, displays the adornment without hovering over its peer"},disableAdornment:{required:!1,tsType:{name:"boolean"},description:""}},composes:["Omit"]};const D={title:"Components/ComboBadge",component:o,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{label:{type:"string",control:"text"},adornment:{type:"string",control:"text"}},args:{}},a={args:{label:"2 📈",adornment:"2+",type:"slideLeft",ariaLabel:"x value increased by 2 points"}},n={args:{label:"2 📈",adornment:"An unreasonably long adornment",type:"slideLeft",ariaLabel:"x value increased by 2 points"}},r={args:{label:"2 📈",adornment:"2+",type:"slideRight",ariaLabel:"x value increased by 2 points"}},t={args:{label:"2 📈",type:"slideRight",ariaLabel:"something"}},s={args:{label:e.jsxs(e.Fragment,{children:["2% ",e.jsx(C,{icon:F,className:"ml-1"})]}),adornment:"2+",type:"inlineSlideLeft",ariaLabel:"x value increased by 2 points",className:"ml-2"},render:l=>e.jsxs("div",{className:"flex justify-center items-center",children:[e.jsx(B,{value:50,color:"primary"})," 8.1K Exposed Principles",e.jsx(o,{...l})]})};var p,c,u;a.parameters={...a.parameters,docs:{...(p=a.parameters)==null?void 0:p.docs,source:{originalSource:`{
  args: {
    label: '2 📈',
    adornment: '2+',
    type: 'slideLeft',
    ariaLabel: 'x value increased by 2 points'
  }
}`,...(u=(c=a.parameters)==null?void 0:c.docs)==null?void 0:u.source}}};var g,b,f;n.parameters={...n.parameters,docs:{...(g=n.parameters)==null?void 0:g.docs,source:{originalSource:`{
  args: {
    label: '2 📈',
    adornment: 'An unreasonably long adornment',
    type: 'slideLeft',
    ariaLabel: 'x value increased by 2 points'
  }
}`,...(f=(b=n.parameters)==null?void 0:b.docs)==null?void 0:f.source}}};var y,x,h;r.parameters={...r.parameters,docs:{...(y=r.parameters)==null?void 0:y.docs,source:{originalSource:`{
  args: {
    label: '2 📈',
    adornment: '2+',
    type: 'slideRight',
    ariaLabel: 'x value increased by 2 points'
  }
}`,...(h=(x=r.parameters)==null?void 0:x.docs)==null?void 0:h.source}}};var v,L,S;t.parameters={...t.parameters,docs:{...(v=t.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    label: '2 📈',
    type: 'slideRight',
    ariaLabel: 'something'
  }
}`,...(S=(L=t.parameters)==null?void 0:L.docs)==null?void 0:S.source}}};var A,R,N;s.parameters={...s.parameters,docs:{...(A=s.parameters)==null?void 0:A.docs,source:{originalSource:`{
  args: {
    label: <>
                2% <FontAwesomeIcon icon={faArrowTrendUp} className='ml-1' />
            </>,
    adornment: '2+',
    type: 'inlineSlideLeft',
    ariaLabel: 'x value increased by 2 points',
    className: 'ml-2'
  },
  render: props => {
    return <div className='flex justify-center items-center'>
                <RadialGauge value={50} color='primary' /> 8.1K Exposed Principles
                <ComboBadge {...props} />
            </div>;
  }
}`,...(N=(R=s.parameters)==null?void 0:R.docs)==null?void 0:N.source}}};const H=["SimpleImpl","ExtendedAdornment","SlideRight","NoAdornment","InlineAdornmentFigmaSpec"];export{n as ExtendedAdornment,s as InlineAdornmentFigmaSpec,t as NoAdornment,a as SimpleImpl,r as SlideRight,H as __namedExportsOrder,D as default};
