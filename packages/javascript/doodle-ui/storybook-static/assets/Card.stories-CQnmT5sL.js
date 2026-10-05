import{j as e}from"./jsx-runtime-BTJTZTIL.js";import{F as y,n as F}from"./index.es-C0EmWI2V.js";import{B}from"./Badge-ldcobJiH.js";import{B as k}from"./Button-DTtPuz4n.js";import{C as t,a as i,d,e as j,b as o,c as v}from"./Card-BdBAz3fL.js";import{c as w}from"./createLucideIcon-rMs-1iam.js";import"./index-ChsGqxH_.js";import"./_commonjsHelpers-CqkleIqs.js";import"./index-2AqrBsEA.js";import"./index-Bb4qSo10.js";import"./utils-BIldDEfN.js";import"./Button.styles-DqX4pzJt.js";import"./useRenderElement--ncz4fp5.js";import"./Typography-BMbK6Q6M.js";/**
 * @license lucide-react v0.436.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const T=w("ChevronUp",[["path",{d:"m18 15-6-6-6 6",key:"153udz"}]]),O={title:"Components/Card",component:t,parameters:{layout:"centered"},tags:["autodocs"],argTypes:{},args:{}},n={render:r=>e.jsxs(t,{...r,children:[e.jsxs(i,{children:[e.jsx(d,{children:"Card Title"}),e.jsx(j,{children:"Card Description"})]}),e.jsx(o,{children:e.jsx("p",{children:"Card Content"})}),e.jsx(v,{children:e.jsx("p",{children:"Card Footer"})})]})},a={render:r=>{const N=[{title:"Your call has been confirmed.",description:"1 hour ago"},{title:"You have a new message!",description:"1 hour ago"},{title:"Your subscription is expiring soon!",description:"2 hours ago"}];return e.jsxs(t,{...r,children:[e.jsxs(i,{children:[e.jsx(d,{children:"Notifications"}),e.jsx(j,{children:"You have 3 unread messages."})]}),e.jsx(o,{className:"grid gap-4",children:e.jsx("div",{children:N.map((c,b)=>e.jsxs("div",{className:"mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0",children:[e.jsx("span",{className:"flex h-2 w-2 translate-y-1 rounded-full bg-sky-500"}),e.jsxs("div",{className:"space-y-1",children:[e.jsx("p",{className:"text-sm font-medium leading-none",children:c.title}),e.jsx("p",{className:"text-sm text-muted-foreground",children:c.description})]})]},b))})}),e.jsx(v,{children:e.jsx("div",{className:"w-full text-center",children:e.jsx(k,{children:"Mark all as read"})})})]})}},s={render:r=>e.jsx(t,{...r,children:e.jsxs("div",{className:"flex gap-12",children:[e.jsxs("div",{children:[e.jsx(i,{children:e.jsx(d,{children:"Findings"})}),e.jsxs(o,{children:[e.jsx("div",{className:"text-4xl font-light mb-3",children:"20K"}),e.jsxs("div",{className:"flex items-center justify-center gap-1",children:[e.jsx(B,{label:"10",icon:e.jsx(T,{}),color:"green"}),e.jsx("span",{className:"text-gray-400",children:"vs last month"})]})]})]}),e.jsx("div",{className:"pt-3 pr-3",children:e.jsx("div",{className:"bg-neutral-light-5 dark:bg-neutral-dark-5 w-8 h-8 rounded-full flex items-center justify-center",children:e.jsx(y,{icon:F})})})]})})};var l,m,p;n.parameters={...n.parameters,docs:{...(l=n.parameters)==null?void 0:l.docs,source:{originalSource:`{
  render: args => {
    return <Card {...args}>
                <CardHeader>
                    <CardTitle>Card Title</CardTitle>
                    <CardDescription>Card Description</CardDescription>
                </CardHeader>
                <CardContent>
                    <p>Card Content</p>
                </CardContent>
                <CardFooter>
                    <p>Card Footer</p>
                </CardFooter>
            </Card>;
  }
}`,...(p=(m=n.parameters)==null?void 0:m.docs)==null?void 0:p.source}}};var u,C,x;a.parameters={...a.parameters,docs:{...(u=a.parameters)==null?void 0:u.docs,source:{originalSource:`{
  render: args => {
    const notifications = [{
      title: 'Your call has been confirmed.',
      description: '1 hour ago'
    }, {
      title: 'You have a new message!',
      description: '1 hour ago'
    }, {
      title: 'Your subscription is expiring soon!',
      description: '2 hours ago'
    }];
    return <Card {...args}>
                <CardHeader>
                    <CardTitle>Notifications</CardTitle>
                    <CardDescription>You have 3 unread messages.</CardDescription>
                </CardHeader>
                <CardContent className='grid gap-4'>
                    <div>
                        {notifications.map((notification, index) => <div key={index} className='mb-4 grid grid-cols-[25px_1fr] items-start pb-4 last:mb-0 last:pb-0'>
                                <span className='flex h-2 w-2 translate-y-1 rounded-full bg-sky-500' />
                                <div className='space-y-1'>
                                    <p className='text-sm font-medium leading-none'>{notification.title}</p>
                                    <p className='text-sm text-muted-foreground'>{notification.description}</p>
                                </div>
                            </div>)}
                    </div>
                </CardContent>
                <CardFooter>
                    <div className='w-full text-center'>
                        <Button>Mark all as read</Button>
                    </div>
                </CardFooter>
            </Card>;
  }
}`,...(x=(C=a.parameters)==null?void 0:C.docs)==null?void 0:x.source}}};var g,h,f;s.parameters={...s.parameters,docs:{...(g=s.parameters)==null?void 0:g.docs,source:{originalSource:`{
  render: args => {
    return <Card {...args}>
                <div className='flex gap-12'>
                    <div>
                        <CardHeader>
                            <CardTitle>Findings</CardTitle>
                        </CardHeader>
                        <CardContent>
                            <div className='text-4xl font-light mb-3'>20K</div>
                            <div className='flex items-center justify-center gap-1'>
                                <Badge label='10' icon={<ChevronUp />} color='green' />
                                <span className='text-gray-400'>vs last month</span>
                            </div>
                        </CardContent>
                    </div>
                    <div className='pt-3 pr-3'>
                        <div className='bg-neutral-light-5 dark:bg-neutral-dark-5 w-8 h-8 rounded-full flex items-center justify-center'>
                            <FontAwesomeIcon icon={faGlobe} />
                        </div>
                    </div>
                </div>
            </Card>;
  }
}`,...(f=(h=s.parameters)==null?void 0:h.docs)==null?void 0:f.source}}};const R=["BasicCard","NotificationsCard","BloodHoundUICountCard"];export{n as BasicCard,s as BloodHoundUICountCard,a as NotificationsCard,R as __namedExportsOrder,O as default};
