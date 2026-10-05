import{j as r}from"./jsx-runtime-BTJTZTIL.js";import{S as _}from"./index-TwK_z8r1.js";import{r as t}from"./index-2AqrBsEA.js";import{c as i}from"./utils-BIldDEfN.js";import{C as T}from"./chevron-right-CeRDmBKm.js";import{c as w}from"./createLucideIcon-rMs-1iam.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";/**
 * @license lucide-react v0.436.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const E=w("Ellipsis",[["circle",{cx:"12",cy:"12",r:"1",key:"41hilf"}],["circle",{cx:"19",cy:"12",r:"1",key:"1wjl8i"}],["circle",{cx:"5",cy:"12",r:"1",key:"1pcz8c"}]]),m=t.forwardRef(({...e},a)=>r.jsx("nav",{ref:a,"aria-label":"breadcrumb",...e}));m.displayName="Breadcrumb";const o=t.forwardRef(({className:e,...a},n)=>r.jsx("ol",{ref:n,className:i("flex flex-wrap items-center gap-1.5 break-words text-xs sm:gap-2.5 dark:text-white",e),...a}));o.displayName="BreadcrumbList";const s=t.forwardRef(({className:e,...a},n)=>r.jsx("li",{ref:n,className:i("inline-flex items-center gap-1.5",e),...a}));s.displayName="BreadcrumbItem";const d=t.forwardRef(({asChild:e,className:a,...n},S)=>{const k=e?_:"a";return r.jsx(k,{ref:S,className:i("rounded-sm transition-colors hover:underline focus:outline-none focus-visible:focus-ring focus-visible:underline",a),...n})});d.displayName="BreadcrumbLink";const l=t.forwardRef(({className:e,...a},n)=>r.jsx("span",{ref:n,role:"link","aria-disabled":"true","aria-current":"page",className:i("font-bold text-primary dark:text-secondary-variant-2",e),...a}));l.displayName="BreadcrumbPage";const c=({children:e,className:a,...n})=>r.jsx("li",{role:"presentation","aria-hidden":"true",className:i("[&>svg]:w-3.5 [&>svg]:h-3.5",a),...n,children:e??r.jsx(T,{})});c.displayName="BreadcrumbSeparator";const B=({className:e,...a})=>r.jsxs("span",{role:"presentation","aria-hidden":"true",className:i("flex h-9 w-9 items-center justify-center",e),...a,children:[r.jsx(E,{className:"h-4 w-4"}),r.jsx("span",{className:"sr-only",children:"More"})]});B.displayName="BreadcrumbElipssis";m.__docgenInfo={description:"",methods:[],displayName:"Breadcrumb",props:{separator:{required:!1,tsType:{name:"ReactReactNode",raw:"React.ReactNode"},description:""}}};B.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbElipssis"};s.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbItem"};d.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbLink",props:{asChild:{required:!1,tsType:{name:"boolean"},description:""}}};o.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbList"};l.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbPage"};c.__docgenInfo={description:"",methods:[],displayName:"BreadcrumbSeparator"};const M={title:"Components/Breadcrumb",component:m,tags:["autodocs"]},u={render:()=>r.jsx(m,{children:r.jsxs(o,{children:[r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier"})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier Details"})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(l,{children:"Selector"})})]})})},p={render:()=>r.jsx(m,{children:r.jsxs(o,{children:[r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier"})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier Details"})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(B,{})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(d,{href:"#",children:"Selector"})}),r.jsx(c,{}),r.jsx(s,{children:r.jsx(l,{children:"Selector Details"})})]})})},b={render:()=>r.jsx(m,{children:r.jsxs(o,{children:[r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier"})}),r.jsx(c,{children:r.jsx("span",{className:"text-xs",children:"→"})}),r.jsx(s,{children:r.jsx(d,{href:"#",children:"Tier Details"})}),r.jsx(c,{children:r.jsx("span",{className:"text-xs",children:"→"})}),r.jsx(s,{children:r.jsx(l,{children:"Selector"})})]})})};var x,h,j;u.parameters={...u.parameters,docs:{...(x=u.parameters)==null?void 0:x.docs,source:{originalSource:`{
  render: () => <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier Details</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>Selector</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
}`,...(j=(h=u.parameters)==null?void 0:h.docs)==null?void 0:j.source}}};var f,I,g;p.parameters={...p.parameters,docs:{...(f=p.parameters)==null?void 0:f.docs,source:{originalSource:`{
  render: () => <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier Details</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbEllipsis />
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Selector</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator />
                <BreadcrumbItem>
                    <BreadcrumbPage>Selector Details</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
}`,...(g=(I=p.parameters)==null?void 0:I.docs)==null?void 0:g.source}}};var N,L,y;b.parameters={...b.parameters,docs:{...(N=b.parameters)==null?void 0:N.docs,source:{originalSource:`{
  render: () => <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator children={<span className='text-xs'>→</span>} />
                <BreadcrumbItem>
                    <BreadcrumbLink href='#'>Tier Details</BreadcrumbLink>
                </BreadcrumbItem>
                <BreadcrumbSeparator children={<span className='text-xs'>→</span>} />
                <BreadcrumbItem>
                    <BreadcrumbPage>Selector</BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
}`,...(y=(L=b.parameters)==null?void 0:L.docs)==null?void 0:y.source}}};const O=["Basic","WithEllipsis","CustomSeparator"];export{u as Basic,b as CustomSeparator,p as WithEllipsis,O as __namedExportsOrder,M as default};
