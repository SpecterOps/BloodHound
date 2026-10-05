import{j as o}from"./jsx-runtime-BTJTZTIL.js";import{B as E}from"./Button-DTtPuz4n.js";import{T as a,a as P,b as R,c as C,d as S,e as q}from"./Tooltip-4xw-YYUR.js";import"./index-ChsGqxH_.js";import"./index-Bb4qSo10.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./index-Dv0R_RYs.js";import"./index-Dc_EP5I0.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-BWEBPL6D.js";import"./index-DnVAETpg.js";import"./index-ejpvWIAw.js";import"./index-DSMk1t8m.js";import"./floating-ui.react-dom-D40P0bPq.js";import"./index-DRLFypRQ.js";import"./index-vQUXxau3.js";import"./index-oIeaYTeP.js";import"./AppIcons-Dp2sojKC.js";import"./Card-BdBAz3fL.js";import"./Typography-BMbK6Q6M.js";import"./UserCog-VADywgBc.js";import"./VisuallyHidden-DT_8ZbEK.js";const eo={title:"Components/Tooltip",component:a,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{open:{type:"boolean",control:"boolean"},defaultOpen:{type:"boolean",control:"boolean"},onOpenChange:{type:"function"},delayDuration:{type:"number",control:"number"},contentWidth:{type:"string",control:"select",options:["default","match_trigger","lg","md","sm"]}},args:{}},e={args:{tooltip:"This is a basic tooltip description"}},r={args:{tooltip:"Example of a tooltip in context",triggerProps:{className:"ml-2"}},render:t=>o.jsxs("div",{className:"flex font-bold items-center justify-center",children:["Severity",o.jsx(a,{...t})]})},i={args:{tooltip:"Lorem ipsum dolor sit amet consectetur adipisicing elit. Repudiandae voluptate assumenda eius ipsum ab, eum earum corrupti provident sit quaerat quasi? Repudiandae minima esse nisi nihil nam quam vel ex?",contentWidth:"lg"},render:t=>o.jsx(a,{...t})},n={args:{tooltip:"This tooltip has a custom trigger to demonstrate that you can add a tooltip to most things",contentWidth:"lg"},render:t=>o.jsx(a,{...t,children:o.jsx(E,{children:"Example"})})},s={args:{tooltip:""},render:()=>o.jsx(P,{children:o.jsxs(R,{children:[o.jsx(C,{children:o.jsx("button",{children:"Example"})}),o.jsx(S,{children:o.jsx(q,{children:"This is an example of a deconstructed tooltip. You can control much more than the shipped Tooltip component. However, the shipped tooltip should cover most cases."})})]})})};var p,l,m;e.parameters={...e.parameters,docs:{...(p=e.parameters)==null?void 0:p.docs,source:{originalSource:`{
  args: {
    tooltip: 'This is a basic tooltip description'
  }
}`,...(m=(l=e.parameters)==null?void 0:l.docs)==null?void 0:m.source}}};var c,d,u;r.parameters={...r.parameters,docs:{...(c=r.parameters)==null?void 0:c.docs,source:{originalSource:`{
  args: {
    tooltip: 'Example of a tooltip in context',
    triggerProps: {
      className: 'ml-2'
    }
  },
  render: args => {
    return <div className='flex font-bold items-center justify-center'>
                Severity
                <Tooltip {...args} />
            </div>;
  }
}`,...(u=(d=r.parameters)==null?void 0:d.docs)==null?void 0:u.source}}};var g,h,T;i.parameters={...i.parameters,docs:{...(g=i.parameters)==null?void 0:g.docs,source:{originalSource:`{
  args: {
    tooltip: \`Lorem ipsum dolor sit amet consectetur adipisicing elit. Repudiandae voluptate assumenda eius ipsum ab, eum earum corrupti provident sit quaerat quasi? Repudiandae minima esse nisi nihil nam quam vel ex?\`,
    contentWidth: 'lg'
  },
  render: args => {
    return <Tooltip {...args} />;
  }
}`,...(T=(h=i.parameters)==null?void 0:h.docs)==null?void 0:T.source}}};var x,v,f;n.parameters={...n.parameters,docs:{...(x=n.parameters)==null?void 0:x.docs,source:{originalSource:`{
  args: {
    tooltip: 'This tooltip has a custom trigger to demonstrate that you can add a tooltip to most things',
    contentWidth: 'lg'
  },
  render: args => {
    return <Tooltip {...args}>
                <Button>Example</Button>
            </Tooltip>;
  }
}`,...(f=(v=n.parameters)==null?void 0:v.docs)==null?void 0:f.source}}};var b,j,y;s.parameters={...s.parameters,docs:{...(b=s.parameters)==null?void 0:b.docs,source:{originalSource:`{
  args: {
    tooltip: ''
  },
  render: () => {
    return <TooltipProvider>
                <TooltipRoot>
                    <TooltipTrigger>
                        <button>Example</button>
                    </TooltipTrigger>
                    <TooltipPortal>
                        <TooltipContent>
                            This is an example of a deconstructed tooltip. You can control much more than the shipped
                            Tooltip component. However, the shipped tooltip should cover most cases.
                        </TooltipContent>
                    </TooltipPortal>
                </TooltipRoot>
            </TooltipProvider>;
  }
}`,...(y=(j=s.parameters)==null?void 0:j.docs)==null?void 0:y.source}}};const ro=["BasicTooltip","InlineTooltip","LongTooltip","CustomTrigger","DeconstructedTooltip"];export{e as BasicTooltip,n as CustomTrigger,s as DeconstructedTooltip,r as InlineTooltip,i as LongTooltip,ro as __namedExportsOrder,eo as default};
