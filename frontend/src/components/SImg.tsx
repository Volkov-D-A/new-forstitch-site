import React from 'react';

interface ImageJob {
  el: HTMLImageElement;
  src: string;
  tries: number;
}

const imgQueue: { q: ImageJob[]; active: number; max: number } = { q: [], active: 0, max: 2 };

function pumpImgQueue() {
  while (imgQueue.active < imgQueue.max && imgQueue.q.length) {
    const job = imgQueue.q.shift();
    if (!job || !job.el.isConnected) continue;

    imgQueue.active++;
    let finished = false;
    const finish = () => {
      if (finished) return;
      finished = true;
      clearTimeout(timer);
      imgQueue.active--;
      pumpImgQueue();
    };

    const timer = setTimeout(() => {
      if (job.tries < 2) imgQueue.q.push({ el: job.el, src: job.src, tries: job.tries + 1 });
      finish();
    }, 14000);

    job.el.onload = () => {
      job.el.classList.add('ld');
      finish();
    };
    job.el.onerror = () => {
      if (job.tries < 2) imgQueue.q.push({ el: job.el, src: job.src, tries: job.tries + 1 });
      finish();
    };
    job.el.src = job.src;
  }
}

type SImgProps = Omit<React.ImgHTMLAttributes<HTMLImageElement>, 'src' | 'alt'> & {
  src?: string;
  alt?: string;
};

export function SImg({ src, alt, className, ...rest }: SImgProps) {
  const ref = React.useRef<HTMLImageElement | null>(null);

  React.useEffect(() => {
    const el = ref.current;
    if (!el || !src) return;
    if (el.src === src && el.complete && el.naturalWidth > 0) return;

    imgQueue.q.push({ el, src, tries: 0 });
    pumpImgQueue();
  }, [src]);

  return (
    <img
      ref={ref}
      alt={alt || ''}
      className={'simg' + (className ? ' ' + className : '')}
      {...rest}
    />
  );
}
