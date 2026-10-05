import{j as p}from"./jsx-runtime-BTJTZTIL.js";import{c as F}from"./index-Bb4qSo10.js";import{b as P,c as g,g as A}from"./utils-BIldDEfN.js";import"./index-ChsGqxH_.js";const I=F("flex justify-center items-center rounded-full text-black",{variants:{type:{labeled:"size-[32px] shadow-none w-auto leading-[1]","sm-labeled":"size-[24px] p-1 shadow-none w-auto leading-[1]","sm-circle":"size-4 p-1 drop-shadow bg-neutral-300","md-circle":"size-8 p-2 drop-shadow bg-neutral-300"}}}),W=F("rounded-full",{variants:{outlined:{true:"border-solid border-2 shadow-none bg-transparent"},type:{labeled:"size-max w-32 px-6 py-2.5 border-none text-center tracking-[.25px]","sm-labeled":"size-max px-1 py-1 border-none text-center font-bold min-w-14 w-14 text-sm","sm-circle":"size-full shadow-inner1xl","md-circle":"size-full shadow-inner1xl"}},compoundVariants:[{outlined:!0,type:"md-circle",className:"border-2"},{outlined:!0,type:"sm-circle",className:"border"}]});function T(e){const{className:d,labelClassName:j,color:_="secondary",outlined:a=!1,type:s,label:q="md-circle",...V}=e,m=A(_),u=s==="labeled"||s==="sm-labeled",E=P([!a,{backgroundColor:m}],[!!a,{borderColor:m}],[!!(a&&u),{boxShadow:`inset 0 0 0 2px ${m}`}]);return p.jsx("div",{role:"status",className:g(I({type:s}),d),...V,children:p.jsx("div",{style:E,className:g(W({outlined:a,type:s}),"risk-badge-label",j),children:p.jsx("span",{children:u?q:null})})})}T.__docgenInfo={description:"",methods:[],displayName:"RiskBadge",props:{color:{required:!1,tsType:{name:"union",raw:"ThemeOptions | CustomColorOptions",elements:[{name:"unknown[number]",raw:"(typeof themeOptions)[number]"},{name:"union",raw:"| `#${string}`\n| `rgb(${string})`\n| `rgba(${string})`\n| `hsl(${string})`\n| `hsla(${string})`",elements:[{name:"literal",value:"`#${string}`"},{name:"literal",value:"`rgb(${string})`"},{name:"literal",value:"`rgba(${string})`"},{name:"literal",value:"`hsl(${string})`"},{name:"literal",value:"`hsla(${string})`"}]}]},description:""},outlined:{required:!0,tsType:{name:"boolean"},description:""},label:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:""},labelClassName:{required:!1,tsType:{name:"string"},description:""}},composes:["VariantProps"]};const r={critical:"#B553EC",high:"#FF3838",moderate:"#FFA86D",low:"#FFD64C",resolved:"#2DCCFF",deprecated:"#D9D9D9"},G=Object.entries(r).map(([e,d])=>({label:e,value:e,hex:d})),Q={title:"Components/RiskBadge",component:T,parameters:{layout:"centered",docs:{description:{component:`
RiskBadges are used to communicate **severity of risk** using color to indicate such risk.

### When to use
- To display **risk severity** (e.g., critical, high, moderate, low)
- To show **exposure** (e.g., percentage of a certain asset that is exposed)

RiskBadges are best used within charts, graphs and paired with severity legend.
                `},canvas:{sourceState:"shown"}}},argTypes:{type:{type:"string",control:"select",options:["labeled","sm-labeled","sm-circle","md-circle"]},color:{control:"select",options:G.map(e=>e.value),mapping:r},outlined:{type:"boolean"}},args:{type:"md-circle",color:"primary",outlined:!1,label:"Label"},tags:["autodocs"]},o={args:{type:"md-circle",color:"primary",outlined:!1}},l={args:{type:"labeled",color:r.critical,label:"Critical",outlined:!1}},t={args:{type:"sm-labeled",color:r.moderate,label:"50%",outlined:!1}},n={args:{type:"labeled",color:r.low,label:"Low",outlined:!0}},i={args:{type:"sm-circle",color:"tertiary",outlined:!1}},c={args:{type:"md-circle",color:"primary",outlined:!0}};var b,y,f;o.parameters={...o.parameters,docs:{...(b=o.parameters)==null?void 0:b.docs,source:{originalSource:`{
  args: {
    type: 'md-circle',
    color: 'primary',
    outlined: false
  }
}`,...(f=(y=o.parameters)==null?void 0:y.docs)==null?void 0:f.source}}};var h,w,x;l.parameters={...l.parameters,docs:{...(h=l.parameters)==null?void 0:h.docs,source:{originalSource:`{
  args: {
    type: 'labeled',
    color: colors.critical,
    label: 'Critical',
    outlined: false
  }
}`,...(x=(w=l.parameters)==null?void 0:w.docs)==null?void 0:x.source}}};var C,v,k;t.parameters={...t.parameters,docs:{...(C=t.parameters)==null?void 0:C.docs,source:{originalSource:`{
  args: {
    type: 'sm-labeled',
    color: colors.moderate,
    label: '50%',
    outlined: false
  }
}`,...(k=(v=t.parameters)==null?void 0:v.docs)==null?void 0:k.source}}};var S,R,$;n.parameters={...n.parameters,docs:{...(S=n.parameters)==null?void 0:S.docs,source:{originalSource:`{
  args: {
    type: 'labeled',
    color: colors.low,
    label: 'Low',
    outlined: true
  }
}`,...($=(R=n.parameters)==null?void 0:R.docs)==null?void 0:$.source}}};var N,O,B;i.parameters={...i.parameters,docs:{...(N=i.parameters)==null?void 0:N.docs,source:{originalSource:`{
  args: {
    type: 'sm-circle',
    color: 'tertiary',
    outlined: false
  }
}`,...(B=(O=i.parameters)==null?void 0:O.docs)==null?void 0:B.source}}};var L,z,D;c.parameters={...c.parameters,docs:{...(L=c.parameters)==null?void 0:L.docs,source:{originalSource:`{
  args: {
    type: 'md-circle',
    color: 'primary',
    outlined: true
  }
}`,...(D=(z=c.parameters)==null?void 0:z.docs)==null?void 0:D.source}}};const U=["Default","Labeled","SmallLabeled","OutlinedLabel","Small","Outlined"];export{o as Default,l as Labeled,c as Outlined,n as OutlinedLabel,i as Small,t as SmallLabeled,U as __namedExportsOrder,Q as default};
