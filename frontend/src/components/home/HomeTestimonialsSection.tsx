import React from 'react';
import { SImg, Stitch } from '../index';
import type { Testimonial } from '../../types/site';

interface TestimonialCardProps {
  testimonial: Testimonial;
}

function TestimonialCard({ testimonial }: TestimonialCardProps) {
  return (
    <figure className="tst-card no-figure-margin">
      <blockquote className="tst-text">{testimonial.text}</blockquote>
      <figcaption className="tst-who">
        <SImg src={testimonial.img} alt={testimonial.name} loading="lazy" />
        <div>
          <div className="tst-name">{testimonial.name}</div>
          <div className="tst-role">{testimonial.role}</div>
        </div>
      </figcaption>
    </figure>
  );
}

interface HomeTestimonialsSectionProps {
  testimonials: Testimonial[];
}

export function HomeTestimonialsSection({ testimonials }: HomeTestimonialsSectionProps) {
  return (
    <section className="sec sec-tint" data-screen-label="Главная: отзывы">
      <div className="wrap">
        <div className="sec-head">
          <div>
            <Stitch />
            <h2 className="h-sec">Отзывы вышивальщиц</h2>
          </div>
        </div>
        <div className="tst-grid">
          {testimonials.map((testimonial) => (
            <TestimonialCard key={testimonial.name} testimonial={testimonial} />
          ))}
        </div>
      </div>
    </section>
  );
}
