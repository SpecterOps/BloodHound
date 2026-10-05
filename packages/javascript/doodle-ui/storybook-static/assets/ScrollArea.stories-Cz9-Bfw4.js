import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{S as l}from"./ScrollArea-CHTOExgL.js";import"./index-ChsGqxH_.js";import"./index-2AqrBsEA.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-BufuEEe8.js";import"./index-BRq5TIsn.js";import"./index-TwK_z8r1.js";import"./index-CNm92B1W.js";import"./utils-BIldDEfN.js";const C=Array.from({length:50}).map((o,r,j)=>`v1.2.0-beta.${j.length-r} Lorem ipsum, dolor sit amet consectetur adipisicing elit. Libero quo voluptas quia optio perspiciatis voluptatibus, blanditiis necessitatibus`),k={title:"Components/ScrollArea",component:l,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{type:{control:{type:"select"},options:["auto","hover","always","scroll"],description:"Determines when and how the scrollbars display.",table:{defaultValue:{summary:"auto"}}},scrollbarWidth:{description:"Scrollbar width in px"},thumbHeight:{description:"Thumbnail height in px"},scrollbarColor:{control:"color"},thumbColor:{control:"color"}},render:o=>e.jsx(l,{...o,className:"h-72 w-48 rounded-md border",children:C.map(r=>e.jsx("div",{className:"relative mt-2.5 border-t pt-2.5 text-[13px] leading-[18px] w-full text-nowrap",children:r},r))})},s={args:{type:"auto"},parameters:{docs:{description:{story:"Scrollbars are visible when content is overflowing on the corresponding orientation."}}}},t={args:{type:"hover"},parameters:{docs:{description:{story:"Scrollbars are visible when the user is scrolling along its corresponding orientation and when the user is hovering over the scroll area."}}}},a={args:{type:"always"},parameters:{docs:{description:{story:"Scrollbars are always visible regardless of whether the content is overflowing."}}}},n={args:{type:"scroll"},parameters:{docs:{description:{story:"Scrollbars are visible when the user is scrolling along its corresponding orientation."}}}},i={args:{type:"auto"},parameters:{docs:{description:{story:`<div>
                <p>usage: </p>
                <div><code><ScrollArea</code></div>
                <div><code>scrollbarWidth={14}</code></div>
                <div><code>thumbColor='lime'</code></div>
                <div><code>scrollbarColor='hsla(209, 82%, 64%, 0.35)'></code></div>
                </div>`}}},render:o=>e.jsx("div",{className:"w-full flex items-center justify-center",children:e.jsx(l,{...o,className:"h-72 w-48 rounded-none  border",scrollbarWidth:14,thumbColor:"lime",scrollbarColor:"hsla(209, 82%, 64%, 0.35)",children:C.map(r=>e.jsx("div",{className:"relative mt-2.5 border-t pt-2.5 text-[13px] leading-[18px] w-full text-nowrap",children:r},r))})})};var c,d,p;s.parameters={...s.parameters,docs:{...(c=s.parameters)==null?void 0:c.docs,source:{originalSource:`{
  args: {
    type: 'auto'
  },
  parameters: {
    docs: {
      description: {
        story: 'Scrollbars are visible when content is overflowing on the corresponding orientation.'
      }
    }
  }
}`,...(p=(d=s.parameters)==null?void 0:d.docs)==null?void 0:p.source}}};var m,u,h;t.parameters={...t.parameters,docs:{...(m=t.parameters)==null?void 0:m.docs,source:{originalSource:`{
  args: {
    type: 'hover'
  },
  parameters: {
    docs: {
      description: {
        story: 'Scrollbars are visible when the user is scrolling along its corresponding orientation and when the user is hovering over the scroll area.'
      }
    }
  }
}`,...(h=(u=t.parameters)==null?void 0:u.docs)==null?void 0:h.source}}};var v,g,b;a.parameters={...a.parameters,docs:{...(v=a.parameters)==null?void 0:v.docs,source:{originalSource:`{
  args: {
    type: 'always'
  },
  parameters: {
    docs: {
      description: {
        story: 'Scrollbars are always visible regardless of whether the content is overflowing.'
      }
    }
  }
}`,...(b=(g=a.parameters)==null?void 0:g.docs)==null?void 0:b.source}}};var y,w,S;n.parameters={...n.parameters,docs:{...(y=n.parameters)==null?void 0:y.docs,source:{originalSource:`{
  args: {
    type: 'scroll'
  },
  parameters: {
    docs: {
      description: {
        story: 'Scrollbars are visible when the user is scrolling along its corresponding orientation.'
      }
    }
  }
}`,...(S=(w=n.parameters)==null?void 0:w.docs)==null?void 0:S.source}}};var x,f,A;i.parameters={...i.parameters,docs:{...(x=i.parameters)==null?void 0:x.docs,source:{originalSource:`{
  args: {
    type: 'auto'
  },
  parameters: {
    docs: {
      description: {
        story: \`<div>
                <p>usage: </p>
                <div><code><ScrollArea</code></div>
                <div><code>scrollbarWidth={14}</code></div>
                <div><code>thumbColor='lime'</code></div>
                <div><code>scrollbarColor='hsla(209, 82%, 64%, 0.35)'></code></div>
                </div>\`
      }
    }
  },
  render: args => <div className='w-full flex items-center justify-center'>
            <ScrollArea {...args} className='h-72 w-48 rounded-none  border' scrollbarWidth={14} thumbColor='lime' scrollbarColor='hsla(209, 82%, 64%, 0.35)'>
                {TAGS.map(tag => <div className='relative mt-2.5 border-t pt-2.5 text-[13px] leading-[18px] w-full text-nowrap' key={tag}>
                        {tag}
                    </div>)}
            </ScrollArea>
        </div>
}`,...(A=(f=i.parameters)==null?void 0:f.docs)==null?void 0:A.source}}};const D=["Auto","Hover","Always","Scroll","OverrideStyles"];export{a as Always,s as Auto,t as Hover,i as OverrideStyles,n as Scroll,D as __namedExportsOrder,k as default};
