import{j as a}from"./jsx-runtime-BTJTZTIL.js";import{F as q,l as G,e as H}from"./index.es-C0EmWI2V.js";import{T as i}from"./Toggle-DnXnoEdn.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-TwK_z8r1.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";const $={title:"Components/Toggle",component:i,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{size:{options:["sm","lg"],control:"select",description:"Size of the toggle button."},disabled:{control:"boolean",description:"Disables interaction with the toggle."}},args:{children:"Toggle Button",size:"sm",disabled:!1}},e={args:{children:"Toggle Button"}},o={args:{size:"lg",children:"Toggle Button"}},r={render:c=>a.jsxs(i,{...c,children:[a.jsx(q,{icon:G}),"Toggle Button"]})},t={render:c=>a.jsxs(i,{...c,children:["Toggle Button",a.jsx(q,{icon:H})]})},s={args:{defaultPressed:!0,children:"Toggle Button"}},n={args:{disabled:!0,children:"Toggle Button"}};var g,d,l,p,m;e.parameters={...e.parameters,docs:{...(g=e.parameters)==null?void 0:g.docs,source:{originalSource:`{
  args: {
    children: 'Toggle Button'
  }
}`,...(l=(d=e.parameters)==null?void 0:d.docs)==null?void 0:l.source},description:{story:"Default/Small size toggle button with outline.",...(m=(p=e.parameters)==null?void 0:p.docs)==null?void 0:m.description}}};var u,h,T,f,b;o.parameters={...o.parameters,docs:{...(u=o.parameters)==null?void 0:u.docs,source:{originalSource:`{
  args: {
    size: 'lg',
    children: 'Toggle Button'
  }
}`,...(T=(h=o.parameters)==null?void 0:h.docs)==null?void 0:T.source},description:{story:"Large size toggle button.",...(b=(f=o.parameters)==null?void 0:f.docs)==null?void 0:b.description}}};var B,I,S,y,D;r.parameters={...r.parameters,docs:{...(B=r.parameters)==null?void 0:B.docs,source:{originalSource:`{
  render: args => <Toggle {...args}>
            <FontAwesomeIcon icon={faCircleInfo} />
            Toggle Button
        </Toggle>
}`,...(S=(I=r.parameters)==null?void 0:I.docs)==null?void 0:S.source},description:{story:"Toggle button with an Icon on the Left.",...(D=(y=r.parameters)==null?void 0:y.docs)==null?void 0:D.description}}};var w,x,z,j,L;t.parameters={...t.parameters,docs:{...(w=t.parameters)==null?void 0:w.docs,source:{originalSource:`{
  render: args => <Toggle {...args}>
            Toggle Button
            <FontAwesomeIcon icon={faStar} />
        </Toggle>
}`,...(z=(x=t.parameters)==null?void 0:x.docs)==null?void 0:z.source},description:{story:"Toggle button with an Icon on the Right.",...(L=(j=t.parameters)==null?void 0:j.docs)==null?void 0:L.description}}};var F,P,R,A,C;s.parameters={...s.parameters,docs:{...(F=s.parameters)==null?void 0:F.docs,source:{originalSource:`{
  args: {
    defaultPressed: true,
    children: 'Toggle Button'
  }
}`,...(R=(P=s.parameters)==null?void 0:P.docs)==null?void 0:R.source},description:{story:"The toggle button rendered in a pressed (active) state by default.",...(C=(A=s.parameters)==null?void 0:A.docs)==null?void 0:C.description}}};var E,_,v,O,k;n.parameters={...n.parameters,docs:{...(E=n.parameters)==null?void 0:E.docs,source:{originalSource:`{
  args: {
    disabled: true,
    children: 'Toggle Button'
  }
}`,...(v=(_=n.parameters)==null?void 0:_.docs)==null?void 0:v.source},description:{story:"Disabled toggle button cannot be interacted with.",...(k=(O=n.parameters)==null?void 0:O.docs)==null?void 0:k.description}}};const ee=["Default","Large","IconLeft","IconRight","DefaultPressed","Disabled"];export{e as Default,s as DefaultPressed,n as Disabled,r as IconLeft,t as IconRight,o as Large,ee as __namedExportsOrder,$ as default};
